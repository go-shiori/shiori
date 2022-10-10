package domains

import (
	"context"
	"fmt"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthDomain struct {
	logger *zap.Logger
	db     database.DB
	secret []byte
}

func (d *AuthDomain) GetAccountFromCredentials(ctx context.Context, username, password string) (*model.Account, error) {
	account, _, err := d.db.GetAccount(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("password do not match")
	}

	return &account, nil
}

func (d *AuthDomain) CreateTokenForAccount(account *model.Account) (string, error) {
	claims := jwt.MapClaims{
		"account": account,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(d.secret)
	if err != nil {
		d.logger.Error("error signing token", zap.Error(err))
	}

	return t, err
}

func NewAuthDomain(logger *zap.Logger, secretKey string, db database.DB) AuthDomain {
	return AuthDomain{
		logger: logger,
		db:     db,
		secret: []byte(secretKey),
	}
}
