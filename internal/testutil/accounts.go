package testutil

import (
	"context"
	"time"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

func NewAdminUser(deps *dependencies.Dependencies) (*model.AccountDTO, string, error) {
	account, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
		Username: "admin",
		Password: "admin",
		Owner:    model.Ptr(true),
	})
	if err != nil {
		return nil, "", err
	}

	token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Hour*24*365))
	if err != nil {
		return nil, "", err
	}

	return account, token, nil
}
