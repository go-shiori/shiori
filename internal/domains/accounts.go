package domains

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type AccountsDomain struct {
	deps *dependencies.Dependencies
}

func (d *AccountsDomain) ListAccounts(ctx context.Context) ([]model.AccountDTO, error) {
	accounts, err := d.deps.Database.ListAccounts(ctx, database.ListAccountsOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting accounts: %v", err)
	}

	accountDTOs := []model.AccountDTO{}
	for _, account := range accounts {
		accountDTOs = append(accountDTOs, account.ToDTO())
	}

	return accountDTOs, nil
}

func (d *AccountsDomain) CreateAccount(ctx context.Context, account model.AccountDTO) (*model.AccountDTO, error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), 10)
	if err != nil {
		return nil, fmt.Errorf("error hashing provided password: %w", err)
	}

	storedAccount, err := d.deps.Database.SaveAccount(ctx, model.Account{
		Username: account.Username,
		Password: string(hashedPassword),
		Owner:    account.Owner,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating account: %v", err)
	}

	result := storedAccount.ToDTO()

	return &result, nil
}

func (d *AccountsDomain) DeleteAccount(ctx context.Context, id int) error {
	err := d.deps.Database.DeleteAccount(ctx, model.DBID(id))
	if errors.Is(err, database.ErrNotFound) {
		return model.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("error deleting account: %v", err)
	}

	return nil
}

func (d *AccountsDomain) UpdateAccount(ctx context.Context, account model.AccountDTO) (*model.AccountDTO, error) {
	updatedAccount := model.Account{
		ID: account.ID,
	}

	// Update password as well
	if account.Password != "" {
		// Hash password with bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), 10)
		if err != nil {
			return nil, fmt.Errorf("error hashing provided password: %w", err)
		}
		updatedAccount.Password = string(hashedPassword)
	}

	// TODO

	return nil, nil
}

func NewAccountsDomain(deps *dependencies.Dependencies) model.AccountsDomain {
	return &AccountsDomain{
		deps: deps,
	}
}
