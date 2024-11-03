package domains

import (
	"context"
	"fmt"
	"time"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type AccountsDomain struct {
	deps *dependencies.Dependencies
}

func (d *AccountsDomain) ParseToken(userJWT string) (*model.JWTClaim, error) {
	token, err := jwt.ParseWithClaims(userJWT, &model.JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return d.deps.Config.Http.SecretKey, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "error parsing token")
	}

	if claims, ok := token.Claims.(*model.JWTClaim); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("error obtaining user from JWT claims")
}

func (d *AccountsDomain) CheckToken(ctx context.Context, userJWT string) (*model.Account, error) {
	claims, err := d.ParseToken(userJWT)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if claims.Account.ID > 0 {
		return claims.Account, nil
	}
	return nil, fmt.Errorf("error obtaining user from JWT claims: %w", err)
}

func (d *AccountsDomain) GetAccountFromCredentials(ctx context.Context, username, password string) (*model.Account, error) {
	account, _, err := d.deps.Database.GetAccount(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("username and password do not match")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("username and password do not match")
	}

	return &account, nil
}

func (d *AccountsDomain) CreateTokenForAccount(account *model.Account, expiration time.Time) (string, error) {
	if account == nil {
		return "", fmt.Errorf("account is nil")
	}

	claims := jwt.MapClaims{
		"account": account.ToDTO(),
		"exp":     expiration.UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(d.deps.Config.Http.SecretKey)
	if err != nil {
		d.deps.Log.WithError(err).Error("error signing token")
	}

	return t, err
}

func NewAccountsDomain(deps *dependencies.Dependencies) *AccountsDomain {
	return &AccountsDomain{
		deps: deps,
	}
}
