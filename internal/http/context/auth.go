package context

// UserIsLogged returns a boolean indicating if the user is authenticated or not
func (c *Context) UserIsLogged() bool {
	_, exists := c.Get("account")
	return exists
}
