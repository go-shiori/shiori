package model

import "errors"

var (
	ErrBookmarkNotFound      = errors.New("bookmark not found")
	ErrBookmarkInvalidID     = errors.New("invalid bookmark ID")
	ErrBookmarkAlreadyExists = errors.New("bookmark already exists")
	ErrTagNotFound           = errors.New("tag not found")
	ErrTagAlreadyExists      = errors.New("tag already exists")

	ErrUnauthorized  = errors.New("unauthorized user")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
