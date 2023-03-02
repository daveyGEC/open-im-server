package msg

import (
	"OpenIM/pkg/common/tokenverify"
	"OpenIM/pkg/proto/msg"
	"OpenIM/pkg/proto/sdkws"
	"context"
)

func (m *msgServer) DelMsgList(ctx context.Context, req *sdkws.DelMsgListReq) (*sdkws.DelMsgListResp, error) {
	resp := &sdkws.DelMsgListResp{}
	if _, err := m.MsgDatabase.DelMsgBySeqs(ctx, req.UserID, req.SeqList); err != nil {
		return nil, err
	}
	DeleteMessageNotification(ctx, req.UserID, req.SeqList)
	return resp, nil
}

func (m *msgServer) DelSuperGroupMsg(ctx context.Context, req *msg.DelSuperGroupMsgReq) (*msg.DelSuperGroupMsgResp, error) {
	resp := &msg.DelSuperGroupMsgResp{}
	if err := tokenverify.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	//maxSeq, err := m.MsgDatabase.GetGroupMaxSeq(ctx, req.GroupID)
	//if err != nil {
	//	return nil, err
	//}
	//if err := m.MsgDatabase.SetGroupUserMinSeq(ctx, req.GroupID, maxSeq); err != nil {
	//	return nil, err
	//}
	if err := m.MsgDatabase.DeleteUserSuperGroupMsgsAndSetMinSeq(ctx, req.GroupID, []string{req.UserID}, 0); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *msgServer) ClearMsg(ctx context.Context, req *msg.ClearMsgReq) (*msg.ClearMsgResp, error) {
	resp := &msg.ClearMsgResp{}
	if err := tokenverify.CheckAccessV3(ctx, req.UserID); err != nil {
		return nil, err
	}
	if err := m.MsgDatabase.CleanUpUserMsg(ctx, req.UserID); err != nil {
		return nil, err
	}
	//if err := m.MsgDatabase.DelUserAllSeq(ctx, req.UserID); err != nil {
	//	return nil, err
	//}
	return resp, nil
}