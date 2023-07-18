package model

// Account is the database model for account.
type Account struct {
	ID       int    `db:"id"       json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password,omitempty"`
	Owner    bool   `db:"owner"    json:"owner"`
}

// ToDTO converts Account to AccountDTO.
func (a Account) ToDTO() AccountDTO {
	return AccountDTO{
		ID:       a.ID,
		Username: a.Username,
		Owner:    a.Owner,
	}
}

// AccountDTO is data transfer object for Account.
type AccountDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Owner    bool   `json:"owner"`
}
