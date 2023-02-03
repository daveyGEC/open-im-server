package relation

import (
	"Open_IM/pkg/common/db/table"
	"Open_IM/pkg/common/tracelog"
	"Open_IM/pkg/utils"
	"context"
	"gorm.io/gorm"
)

//var FriendRequestDB *gorm.DB

func NewFriendRequestGorm(db *gorm.DB) *FriendRequestGorm {
	var fr FriendRequestGorm
	fr.DB = db
	return &fr
}

type FriendRequestGorm struct {
	DB *gorm.DB `gorm:"-"`
}

func (f *FriendRequestGorm) Create(ctx context.Context, friends []*table.FriendRequestModel) (err error) {
	defer func() {
		tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "friends", friends)
	}()
	return utils.Wrap(f.DB.Model(&table.FriendRequestModel{}).Create(&friends).Error, "")
}

func (f *FriendRequestGorm) Delete(ctx context.Context, fromUserID, toUserID string) (err error) {
	defer func() {
		tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "fromUserID", fromUserID, "toUserID", toUserID)
	}()
	return utils.Wrap(f.DB.Model(&table.FriendRequestModel{}).Where("from_user_id = ? and to_user_id = ?", fromUserID, toUserID).Delete(&table.FriendRequestModel{}).Error, "")
}

func (f *FriendRequestGorm) UpdateByMap(ctx context.Context, ownerUserID string, args map[string]interface{}) (err error) {
	defer func() {
		tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "ownerUserID", ownerUserID, "args", args)
	}()
	return utils.Wrap(f.DB.Model(&table.FriendRequestModel{}).Where("owner_user_id = ?", ownerUserID).Updates(args).Error, "")
}

func (f *FriendRequestGorm) Update(ctx context.Context, friendRequests []*table.FriendRequestModel, tx ...*gorm.DB) (err error) {
	defer func() {
		tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "friendRequests", friendRequests)
	}()
	return utils.Wrap(getDBConn(f.DB, tx).Model(&table.FriendRequestModel{}).Updates(&friendRequests).Error, "")
}

func (f *FriendRequestGorm) Find(ctx context.Context, ownerUserID string) (friends []*table.FriendRequestModel, err error) {
	defer func() {
		tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "ownerUserID", ownerUserID, "friends", friends)
	}()
	return friends, utils.Wrap(f.DB.Model(&table.FriendRequestModel{}).Where("owner_user_id = ?", ownerUserID).Find(&friends).Error, "")
}

func (f *FriendRequestGorm) Take(ctx context.Context, fromUserID, toUserID string) (friend *table.FriendRequestModel, err error) {
	friend = &table.FriendRequestModel{}
	defer tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "fromUserID", fromUserID, "toUserID", toUserID, "friend", friend)
	return friend, utils.Wrap(f.DB.Model(&table.FriendRequestModel{}).Where("from_user_id = ? and to_user_id", fromUserID, toUserID).Take(friend).Error, "")
}

func (f *FriendRequestGorm) FindToUserID(ctx context.Context, toUserID string, pageNumber, showNumber int32, tx ...*gorm.DB) (friends []*table.FriendRequestModel, total int64, err error) {
	defer func() {
		tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "toUserID", toUserID, "friends", friends)
	}()

	err = getDBConn(f.DB, tx).Model(&table.FriendRequestModel{}).Where("to_user_id = ? ", toUserID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Wrap(err, "")
	}
	err = utils.Wrap(getDBConn(f.DB, tx).Model(&table.FriendRequestModel{}).Where("to_user_id = ? ", toUserID).Limit(int(showNumber)).Offset(int(pageNumber*showNumber)).Find(&friends).Error, "")
	return
}

func (f *FriendRequestGorm) FindFromUserID(ctx context.Context, fromUserID string, pageNumber, showNumber int32, tx ...*gorm.DB) (friends []*table.FriendRequestModel, total int64, err error) {
	defer func() {
		tracelog.SetCtxDebug(ctx, utils.GetSelfFuncName(), err, "fromUserID", fromUserID, "friends", friends)
	}()

	err = getDBConn(f.DB, tx).Model(&table.FriendRequestModel{}).Where("from_user_id = ? ", fromUserID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Wrap(err, "")
	}
	err = utils.Wrap(getDBConn(f.DB, tx).Model(&table.FriendRequestModel{}).Where("from_user_id = ? ", fromUserID).Limit(int(showNumber)).Offset(int(pageNumber*showNumber)).Find(&friends).Error, "")
	return
}
