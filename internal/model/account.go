package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Account is the database representation for account.
type Account struct {
	ID       DBID       `db:"id"       json:"id"`
	Username string     `db:"username" json:"username"`
	Password string     `db:"password" json:"password,omitempty"`
	Owner    bool       `db:"owner"    json:"owner"`
	Config   UserConfig `db:"config"               json:"config"`
}

type UserConfig struct {
	ShowId        bool
	ListMode      bool
	HideThumbnail bool
	HideExcerpt   bool
	Theme         string
	KeepMetadata  bool
	UseArchive    bool
	CreateEbook   bool
	MakePublic    bool
}

func (c *UserConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		json.Unmarshal(v, &c)
		return nil
	case string:
		json.Unmarshal([]byte(v), &c)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func (c UserConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// ToDTO converts Account to AccountDTO.
func (a Account) ToDTO() AccountDTO {
	owner := a.Owner
	config := a.Config

	return AccountDTO{
		ID:       a.ID,
		Username: a.Username,
		Owner:    &owner,
		Config:   &config,
	}
}

// AccountDTO is data transfer object for Account.
type AccountDTO struct {
	ID       DBID        `json:"id"`
	Username string      `json:"username"`
	Password string      `json:"passowrd,omitempty"` // Used only to store, not to retrieve
	Owner    *bool       `json:"owner"`
	Config   *UserConfig `json:"config"`
}

func (adto *AccountDTO) IsOwner() bool {
	return adto.Owner != nil && *adto.Owner
}

func (adto *AccountDTO) IsValidCreate() error {
	if adto.Username == "" {
		return NewValidationError("username", "username should not be empty")
	}

	if adto.Password == "" {
		return NewValidationError("password", "password should not be empty")
	}

	return nil
}

func (adto *AccountDTO) IsValidUpdate() error {
	if adto.Username == "" && adto.Password == "" && adto.Owner == nil && adto.Config == nil {
		return NewValidationError("account", "no fields to update")
	}

	return nil
}

type JWTClaim struct {
	jwt.RegisteredClaims

	Account *Account
}
