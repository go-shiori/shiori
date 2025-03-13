package model

import "errors"

var (
	ErrBookmarkNotFound  = errors.New("bookmark not found")
	ErrBookmarkInvalidID = errors.New("invalid bookmark ID")
	ErrTagNotFound       = errors.New("tag not found")

	ErrUnauthorized  = errors.New("unauthorized user")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
