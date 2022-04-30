package errors

import "errors"

var (
	ErrUserExisted     = errors.New("")
	ErrUserNotExisted  = errors.New("")
	ErrNotEnterRoom    = errors.New("")
	ErrUserNotLogin    = errors.New("")
	ErrUserOnline      = errors.New("")
	ErrUnknownGmCmd    = errors.New("")
	ErrInvalidRoomID   = errors.New("")
	ErrRoomIDNotExists = errors.New("")
)
