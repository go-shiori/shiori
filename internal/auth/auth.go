package auth

import (
	"database/sql"
	"fmt"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/env"
	"golang.org/x/crypto/bcrypt"
)

// Status is the authentication status returned by Check
type Status int

const (
	// Unauthorized username or wrong password
	Unauthorized Status = iota
	// Visitor means that the user is authorized as a visitor
	Visitor
	// Owner means that the user is authorized as an owner
	Owner
)

var (
	ldapAuth *LDAPAuth
)

func init() {
	if env.GetEnvBool("SHIORI_AUTH_LDAP", false) {
		tmp, err := NewLDAPAuth(LDAPSettings{
			Host:                 env.GetEnvString("SHIORI_AUTH_LDAP_HOST", "ldap"),
			Port:                 int(env.GetEnvInt64("SHIORI_AUTH_LDAP_PORT", 389)),
			StartTLS:             env.GetEnvBool("SHIORI_AUTH_LDAP_TLS_ENABLED", true),
			SkipCertificateVerif: env.GetEnvBool("SHIORI_AUTH_LDAP_TLS_SKIP_VERIF", true),
			ThrustedCertificates: env.GetEnvStringList("SHIORI_AUTH_LDAP_TLS_THRUSTED_CERTIFICATES", []string{}),
			UserGroupFilter: env.GetEnvString(
				"SHIORI_AUTH_LDAP_SEARCH_FILTER",
				"(&(|(mail={{.Login}})(sAMAccountName={{.Login}}))(memberOf={{.Group}}))",
			),
			BaseDN:         env.GetEnvString("SHIORI_AUTH_LDAP_SEARCH_BASE", ""),
			BindDN:         env.GetEnvString("SHIORI_AUTH_LDAP_BIND_USERDN", ""),
			BindDNPassword: env.GetEnvString("SHIORI_AUTH_LDAP_BIND_PASSWORD", ""),
		})
		if err == nil {
			ldapAuth = &tmp
		}
	}
}

// Check username, password with configured auth methods
func Check(username string, password string, db database.DB) (Status, string) {

	if ldapAuth != nil {
		loginField := env.GetEnvString("SHIORI_AUTH_LDAP_LOGIN_FIELD", "sAMAccountName")
		ownerGroup := env.GetEnvString("SHIORI_AUTH_LDAP_OWNER_GROUP", "")
		visitorGroup := env.GetEnvString("SHIORI_AUTH_LDAP_VISITOR_GROUP", "")
		oDN, oLogin, oErr := ldapAuth.Search(
			username,
			ownerGroup,
			loginField,
		)
		vDN, vLogin, vErr := ldapAuth.Search(
			username,
			visitorGroup,
			loginField,
		)

		if oErr == nil {
			fmt.Printf("LDAP: owner found: %s\n", oDN)
			if ldapAuth.VerifyDN(oDN, password) == nil {
				return Owner, oLogin
			}
		} else if vErr == nil {
			fmt.Printf("LDAP: visitor found: %s\n", vDN)
			if ldapAuth.VerifyDN(vDN, password) == nil {
				return Visitor, vLogin
			}
		}
		fmt.Printf("LDAP: not found (%v, %v)\n", oErr, vErr)
	}

	defaultUser := env.GetEnvString("SHIORI_DEFAULT_USER", "shiori")
	defaultPassword := env.GetEnvString("SHIORI_DEFAULT_PASSWORD", "gopher")

	// Check if user's database is empty or there are no owner.
	// If yes, and user uses default account, let him in.
	searchOptions := database.GetAccountsOptions{
		Owner: true,
	}

	accounts, err := db.GetAccounts(searchOptions)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if len(accounts) == 0 && username == defaultUser && password == defaultPassword {
		return Owner, defaultUser
	}

	// Get account data from database
	account, exist := db.GetAccount(username)
	hash := ""
	if exist {
		hash = account.Password
	}

	// Compare password with database
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if hash == "" || err != nil {
		return Unauthorized, username
	}

	// If login request is as owner, make sure this account is owner
	if account.Owner {
		return Owner, username
	}
	return Visitor, username

}
