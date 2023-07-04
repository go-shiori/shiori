package context

import "github.com/go-shiori/shiori/internal/model"

// UserIsLogged returns a boolean indicating if the user is authenticated or not
func (c *Context) UserIsLogged() bool {
	_, exists := c.Get(model.ContextAccountKey)
	return exists
}

func (c *Context) GetAccount() *model.Account {
	if c.account == nil && c.UserIsLogged() {
		c.account = c.MustGet(model.ContextAccountKey).(*model.Account)
	}

	return c.account
}
