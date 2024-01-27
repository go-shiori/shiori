package domains

import (
	"context"
	"fmt"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

type AccountsDomain struct {
	deps *dependencies.Dependencies
}

func (d *AccountsDomain) ListAccounts(ctx context.Context) ([]model.AccountDTO, error) {
	accounts, err := d.deps.Database.GetAccounts(ctx, database.GetAccountsOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting accounts: %v", err)
	}

	accountDTOs := []model.AccountDTO{}
	for _, account := range accounts {
		accountDTOs = append(accountDTOs, account.ToDTO())
	}

	return accountDTOs, nil
}

func (d *AccountsDomain) CreateAccount(ctx context.Context, account model.Account) (*model.AccountDTO, error) {
	storedAccount, err := d.deps.Database.SaveAccount(ctx, model.Account{
		Username: account.Username,
		Password: account.Password,
		Owner:    account.Owner,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating account: %v", err)
	}

	// FIXME
	result := storedAccount.ToDTO()

	return &result, nil
}

func NewAccountsDomain(deps *dependencies.Dependencies) model.AccountsDomain {
	return &AccountsDomain{
		deps: deps,
	}
}
