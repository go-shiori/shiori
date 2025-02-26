package testutil

import (
	"context"
	"time"

	"github.com/go-shiori/shiori/internal/model"
)

// NewAdminUser creates a new admin user and returns its account and token.
// Use this when testing the API endpoints that require admin authentication to
// generate the user and obtain a token that can be easily added as `WithAuthToken()`
// option in the request.
func NewAdminUser(deps model.Dependencies) (*model.AccountDTO, string, error) {
	account, err := deps.Domains().Accounts().CreateAccount(context.TODO(), model.AccountDTO{
		Username: "admin",
		Password: "admin",
		Owner:    model.Ptr(true),
	})
	if err != nil {
		return nil, "", err
	}

	token, err := deps.Domains().Auth().CreateTokenForAccount(account, time.Now().Add(time.Hour*24*365))
	if err != nil {
		return nil, "", err
	}

	return account, token, nil
}
