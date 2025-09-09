package webcontext

import (
	"context"

	"github.com/go-shiori/shiori/internal/model"
)

// UserIsLogged returns a boolean indicating if the user is authenticated or not
func (c *WebContext) UserIsLogged() bool {
	return c.GetAccount() != nil
}

// GetAccount retrieves the account from the request context
func (c *WebContext) GetAccount() *model.AccountDTO {
	if acc := c.request.Context().Value(accountKey); acc != nil {
		return acc.(*model.AccountDTO)
	}
	return nil
}

// SetAccount stores the account in the request context
func (c *WebContext) SetAccount(account *model.AccountDTO) {
	ctx := WithAccount(c.request.Context(), account)
	c.request = c.request.WithContext(ctx)
}

// WithAccount creates a new context with the account
func WithAccount(ctx context.Context, account *model.AccountDTO) context.Context {
	return context.WithValue(ctx, accountKey, account)
}
