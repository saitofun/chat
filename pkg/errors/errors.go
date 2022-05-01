package errors

import "errors"

var (
	ErrUserExisted     = errors.New("用户已存在, 请勿重复创建")
	ErrUserNotExisted  = errors.New("用户不存在, 请先创建用户")
	ErrUserNotLogin    = errors.New("用户尚未登陆, 请先登录")
	ErrUserOnline      = errors.New("用户已经登陆, 请勿重复登陆")
	ErrNotEnterRoom    = errors.New("尚未进入房间, 请选择房间或创建房间")
	ErrUnknownGmCmd    = errors.New("未知指令")
	ErrInvalidRoomID   = errors.New("非法的房间号")
	ErrRoomIDExists    = errors.New("房间已存在")
	ErrRoomIDNotExists = errors.New("房间号不存在")
)
