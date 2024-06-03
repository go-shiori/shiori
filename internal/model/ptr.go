package model

func Ptr[t any](a t) *t {
	return &a
}
