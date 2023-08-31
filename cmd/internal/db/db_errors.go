package db

import "errors"

var (
	ErrSegmentNotFound       = errors.New("url not found")
	ErrSegmentAlreadyExists  = errors.New("url already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyHasSegment = errors.New("user already has such segment")
)
