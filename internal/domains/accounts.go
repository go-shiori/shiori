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
	accounts, err := d.deps.Database().ListAccounts(ctx, model.DBListAccountsOptions{})
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
	if err := account.IsValidCreate(); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing provided password: %w", err)
	}

	acc := model.Account{
		Username: account.Username,
		Password: string(hashedPassword),
	}
	if account.Owner != nil {
		acc.Owner = *account.Owner
	}
	if account.Config != nil {
		acc.Config = *account.Config
	}

	storedAccount, err := d.deps.Database().CreateAccount(ctx, acc)
	if errors.Is(err, database.ErrAlreadyExists) {
		return nil, model.ErrAlreadyExists
	}

	if err != nil {
		return nil, fmt.Errorf("error creating account: %v", err)
	}

	result := storedAccount.ToDTO()

	return &result, nil
}

func (d *AccountsDomain) DeleteAccount(ctx context.Context, id int) error {
	err := d.deps.Database().DeleteAccount(ctx, model.DBID(id))
	if errors.Is(err, database.ErrNotFound) {
		return model.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("error deleting account: %v", err)
	}

	return nil
}

func (d *AccountsDomain) UpdateAccount(ctx context.Context, account model.AccountDTO) (*model.AccountDTO, error) {
	if err := account.IsValidUpdate(); err != nil {
		return nil, err
	}

	// Get account from database
	storedAccount, _, err := d.deps.Database().GetAccount(ctx, account.ID)
	if errors.Is(err, database.ErrNotFound) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting account for update: %w", err)
	}

	if account.Password != "" {
		// Hash password with bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), 10)
		if err != nil {
			return nil, fmt.Errorf("error hashing provided password: %w", err)
		}
		storedAccount.Password = string(hashedPassword)
	}

	if account.Username != "" {
		storedAccount.Username = account.Username
	}

	if account.Owner != nil {
		storedAccount.Owner = *account.Owner
	}

	if account.Config != nil {
		storedAccount.Config = *account.Config
	}

	// Save updated account
	err = d.deps.Database().UpdateAccount(ctx, *storedAccount)
	if errors.Is(err, database.ErrAlreadyExists) {
		return nil, model.ErrAlreadyExists
	}

	if err != nil {
		return nil, fmt.Errorf("error updating account: %w", err)
	}

	// Get updated account from database
	updatedAccount, _, err := d.deps.Database().GetAccount(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting updated account: %w", err)
	}

	account = updatedAccount.ToDTO()

	return &account, nil
}

func NewAccountsDomain(deps *dependencies.Dependencies) model.AccountsDomain {
	return &AccountsDomain{
		deps: deps,
	}
}
