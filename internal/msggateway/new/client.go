package new

import (
	"OpenIM/pkg/common/constant"
	"OpenIM/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"runtime/debug"
	"sync"
)

const (
	// MessageText is for UTF-8 encoded text messages like JSON.
	MessageText = iota + 1
	// MessageBinary is for binary messages like protobufs.
	MessageBinary
	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

type Client struct {
	w              *sync.Mutex
	conn           LongConn
	platformID     int
	isCompress     bool
	userID         string
	isBackground   bool
	connID         string
	onlineAt       int64 // 上线时间戳（毫秒）
	handler        MessageHandler
	unregisterChan chan *Client
	compressor     Compressor
	encoder        Encoder
	validate       *validator.Validate
	closed         bool
}

func newClient(ctx *UserConnContext, conn LongConn, isCompress bool, compressor Compressor, encoder Encoder,
	handler MessageHandler, unregisterChan chan *Client, validate *validator.Validate) *Client {
	return &Client{
		w:              new(sync.Mutex),
		conn:           conn,
		platformID:     utils.StringToInt(ctx.GetPlatformID()),
		isCompress:     isCompress,
		userID:         ctx.GetUserID(),
		compressor:     compressor,
		encoder:        encoder,
		connID:         ctx.GetConnID(),
		onlineAt:       utils.GetCurrentTimestampByMill(),
		handler:        handler,
		unregisterChan: unregisterChan,
		validate:       validate,
	}
}
func (c *Client) ResetClient(ctx *UserConnContext, conn LongConn, isCompress bool, compressor Compressor, encoder Encoder,
	handler MessageHandler, unregisterChan chan *Client, validate *validator.Validate) {
	c.w = new(sync.Mutex)
	c.conn = conn
	c.platformID = utils.StringToInt(ctx.GetPlatformID())
	c.isCompress = isCompress
	c.userID = ctx.GetUserID()
	c.compressor = compressor
	c.encoder = encoder
	c.connID = ctx.GetConnID()
	c.onlineAt = utils.GetCurrentTimestampByMill()
	c.handler = handler
	c.unregisterChan = unregisterChan
	c.validate = validate
}
func (c *Client) readMessage() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("socket have panic err:", r, string(debug.Stack()))
		}
		//c.close()
	}()
	//var returnErr error
	for {
		messageType, message, returnErr := c.conn.ReadMessage()
		if returnErr != nil {
			break
		}
		if c.closed == true { //连接刚置位已经关闭，但是协程还没退出的场景
			break
		}
		switch messageType {
		case PingMessage:
		case PongMessage:
		case CloseMessage:
			return
		case MessageText:
		case MessageBinary:
			if len(message) == 0 {
				continue
			}
			returnErr = c.handleMessage(message)
			if returnErr != nil {
				break
			}

		}
	}

}
func (c *Client) handleMessage(message []byte) error {
	if c.isCompress {
		var decompressErr error
		message, decompressErr = c.compressor.DeCompress(message)
		if decompressErr != nil {
			return utils.Wrap(decompressErr, "")
		}
	}
	var binaryReq Req
	err := c.encoder.Decode(message, &binaryReq)
	if err != nil {
		return utils.Wrap(err, "")
	}
	if err := c.validate.Struct(binaryReq); err != nil {
		return utils.Wrap(err, "")
	}
	if binaryReq.SendID != c.userID {
		return errors.New("exception conn userID not same to req userID")
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, CONN_ID, c.connID)
	ctx = context.WithValue(ctx, OPERATION_ID, binaryReq.OperationID)
	ctx = context.WithValue(ctx, COMMON_USERID, binaryReq.SendID)
	ctx = context.WithValue(ctx, PLATFORM_ID, c.platformID)
	var messageErr error
	var resp []byte
	switch binaryReq.ReqIdentifier {
	case constant.WSGetNewestSeq:
		resp, messageErr = c.handler.GetSeq(ctx, binaryReq)
	case constant.WSSendMsg:
		resp, messageErr = c.handler.SendMessage(ctx, binaryReq)
	case constant.WSSendSignalMsg:
		resp, messageErr = c.handler.SendSignalMessage(ctx, binaryReq)
	case constant.WSPullMsgBySeqList:
		resp, messageErr = c.handler.PullMessageBySeqList(ctx, binaryReq)
	case constant.WsLogoutMsg:
		resp, messageErr = c.handler.UserLogout(ctx, binaryReq)
	case constant.WsSetBackgroundStatus:
		resp, messageErr = c.setAppBackgroundStatus(ctx, binaryReq)
	default:
		return errors.New(fmt.Sprintf("ReqIdentifier failed,sendID:%d,msgIncr:%s,reqIdentifier:%s", binaryReq.SendID, binaryReq.MsgIncr, binaryReq.ReqIdentifier))
	}
	c.replyMessage(&binaryReq, messageErr, resp)
	return nil

}
func (c *Client) setAppBackgroundStatus(ctx context.Context, req Req) ([]byte, error) {
	resp, isBackground, messageErr := c.handler.SetUserDeviceBackground(ctx, req)
	if messageErr != nil {
		return nil, messageErr
	}
	c.isBackground = isBackground
	//todo callback
	return resp, nil

}
func (c *Client) close() {
	c.w.Lock()
	defer c.w.Unlock()
	c.conn.Close()
	c.unregisterChan <- c

}
func (c *Client) replyMessage(binaryReq *Req, err error, resp []byte) {
	mReply := Resp{
		ReqIdentifier: binaryReq.ReqIdentifier,
		MsgIncr:       binaryReq.MsgIncr,
		OperationID:   binaryReq.OperationID,
		Data:          resp,
	}
	_ = c.writeMsg(mReply)
}

func (c *Client) writeMsg(resp Resp) error {
	c.w.Lock()
	defer c.w.Unlock()
	if c.closed == true {
		return nil
	}
	encodedBuf := bufferPool.Get().([]byte)
	resultBuf := bufferPool.Get().([]byte)
	encodeBuf, err := c.encoder.Encode(resp)
	if err != nil {
		return utils.Wrap(err, "")
	}
	_ = c.conn.SetWriteTimeout(60)
	if c.isCompress {
		var compressErr error
		resultBuf, compressErr = c.compressor.Compress(encodeBuf)
		if compressErr != nil {
			return utils.Wrap(compressErr, "")
		}
		return c.conn.WriteMessage(MessageBinary, resultBuf)
	} else {
		return c.conn.WriteMessage(MessageBinary, encodedBuf)
	}
}