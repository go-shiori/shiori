package domains

import (
	"context"
	"fmt"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthDomain struct {
	logger *zap.Logger
	db     database.DB
}

func (d *AuthDomain) Login(ctx context.Context, username, password string) (*model.Account, error) {
	account, _, err := d.db.GetAccount(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("password do not match")
	}

	return &account, nil
}

func NewAuthDomain(logger *zap.Logger, db database.DB) AuthDomain {
	return AuthDomain{
		logger: logger,
		db:     db,
	}
}
