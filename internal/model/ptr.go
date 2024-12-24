package model

// Ptr returns a pointer to the value passed as argument.
func Ptr[t any](a t) *t {
	return &a
}
