package domains

import (
	"context"
	"fmt"
	"time"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthDomain struct {
	deps *dependencies.Dependencies
}

type JWTClaim struct {
	jwt.RegisteredClaims

	Account *model.AccountDTO
}

func (d *AuthDomain) CheckToken(ctx context.Context, userJWT string) (*model.AccountDTO, error) {
	token, err := jwt.ParseWithClaims(userJWT, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return d.deps.Config().Http.SecretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaim); ok && token.Valid {
		if claims.Account.ID > 0 {
			return claims.Account, nil
		}

		return claims.Account, nil
	}
	return nil, fmt.Errorf("error obtaining user from JWT claims")
}

func (d *AuthDomain) GetAccountFromCredentials(ctx context.Context, username, password string) (*model.AccountDTO, error) {
	accounts, err := d.deps.Database().ListAccounts(ctx, model.DBListAccountsOptions{
		Username:     username,
		WithPassword: true,
	})
	if err != nil {
		return nil, fmt.Errorf("username or password do not match")
	}

	if len(accounts) != 1 {
		return nil, fmt.Errorf("username or password do not match")
	}

	account := accounts[0]

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("username or password do not match")
	}

	return model.Ptr(account.ToDTO()), nil
}

func (d *AuthDomain) CreateTokenForAccount(account *model.AccountDTO, expiration time.Time) (string, error) {
	if account == nil {
		return "", fmt.Errorf("account is nil")
	}

	claims := jwt.MapClaims{
		"account": account,
		"exp":     expiration.UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(d.deps.Config().Http.SecretKey)
	if err != nil {
		d.deps.Logger().WithError(err).Error("error signing token")
	}

	return t, err
}

func NewAuthDomain(deps *dependencies.Dependencies) *AuthDomain {
	return &AuthDomain{
		deps: deps,
	}
}
