package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	ShowId        bool `json:"ShowId"`
	ListMode      bool `json:"ListMode"`
	HideThumbnail bool `json:"HideThumbnail"`
	HideExcerpt   bool `json:"HideExcerpt"`
	NightMode     bool `json:"NightMode"`
	KeepMetadata  bool `json:"KeepMetadata"`
	UseArchive    bool `json:"UseArchive"`
	CreateEbook   bool `json:"CreateEbook"`
	MakePublic    bool `json:"MakePublic"`
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
	return AccountDTO{
		ID:       a.ID,
		Username: a.Username,
		Owner:    a.Owner,
		Config:   a.Config,
	}
}

// AccountDTO is data transfer object for Account.
type AccountDTO struct {
	ID       DBID       `json:"id"`
	Username string     `json:"username"`
	Password string     `json:"-"` // Used only to store, not to retrieve
	Owner    bool       `json:"owner"`
	Config   UserConfig `json:"config"`
}
