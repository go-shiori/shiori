package model

import "errors"

var (
	ErrBookmarkNotFound  = errors.New("bookmark not found")
	ErrBookmarkInvalidID = errors.New("invalid bookmark ID")
	ErrUnauthorized      = errors.New("unauthorized user")
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
)
