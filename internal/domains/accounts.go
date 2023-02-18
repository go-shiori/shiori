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

type AccountsDomain struct {
	logger *zap.Logger
	db     database.DB
	secret []byte
}

func (d *AccountsDomain) GetAccountFromCredentials(ctx context.Context, username, password string) (*model.Account, error) {
	account, _, err := d.db.GetAccount(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("username and password do not match")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("username and password do not match")
	}

	return &account, nil
}

func (d *AccountsDomain) CreateTokenForAccount(account *model.Account, expiration time.Time) (string, error) {
	claims := jwt.MapClaims{
		"account": account,
		"exp":     expiration.UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(d.secret)
	if err != nil {
		d.logger.Error("error signing token", zap.Error(err))
	}

	return t, err
}

func NewAccountsDomain(logger *zap.Logger, secretKey string, db database.DB) AccountsDomain {
	return AccountsDomain{
		logger: logger,
		db:     db,
		secret: []byte(secretKey),
	}
}
