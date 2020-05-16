package auth

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/go-ldap/ldap"
)

// LDAPAuth class
type LDAPAuth struct {
	settings LDAPSettings
	certs    *x509.CertPool
}

// LDAPSettings used for auth object creation
type LDAPSettings struct {
	Host                 string   // ldap host use full hostname if tls is used
	Port                 int      // ldap port default is 389
	StartTLS             bool     // Start TLS, Disable if not supported by server but credentials will transit without any encryption
	SkipCertificateVerif bool     // Skip certificate verification only use for debug purpose
	ThrustedCertificates []string // List of thrusted CA and certificates
	UserGroupFilter      string   // Filter used to search provided user{{.Login}} & group{{.Group}}
	BaseDN               string   // Base DN for users
	BindDN               string   // DN used to bind for search operations
	BindDNPassword       string   // DN credential used to bind for search operations
}

// NewLDAPAuth returns a ldap auth object from given setting
func NewLDAPAuth(settings LDAPSettings) (LDAPAuth, error) {
	la := LDAPAuth{
		settings: settings,
	}

	la.certs = x509.NewCertPool()
	for _, cert := range la.settings.ThrustedCertificates {
		if data, err := ioutil.ReadFile(cert); err == nil {
			if !la.certs.AppendCertsFromPEM(data) {
				log.Println("ERROR: LDAP Unable to load certificate " + cert)
			}
		} else {
			log.Println("ERROR: LDAP Unable to read certificate " + cert + " " + err.Error())
		}

	}
	return la, nil
}

func (la *LDAPAuth) connect() (*ldap.Conn, error) {
	// log.Println("LDAP:connect: Dial")
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", la.settings.Host, la.settings.Port))
	if err != nil {
		return nil, err
	}

	if la.settings.StartTLS {
		// Reconnect with TLS
		var tlsConfig tls.Config
		if la.settings.SkipCertificateVerif {
			log.Println("WARNING: LDAP LDAPAuth with TLS without certificate verification")
			tlsConfig = tls.Config{InsecureSkipVerify: true}
		} else {
			tlsConfig = tls.Config{
				ServerName:         la.settings.Host,
				InsecureSkipVerify: false,
				RootCAs:            la.certs,
			}
		}
		// log.Println("LDAP:connect: StartTLS")
		err = l.StartTLS(&tlsConfig)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("WARNING: LDAP LDAPAuth without TLS")
	}

	// Bind with read only user
	// log.Println("LDAP:connect: Bind as", la.settings.BindDN)
	err = l.Bind(la.settings.BindDN, la.settings.BindDNPassword)
	if err != nil {
		return nil, err
	}
	// log.Println("LDAP:connect: Done")
	return l, nil
}

func (la *LDAPAuth) search(l *ldap.Conn, username string, group string, loginField string) (string, string, error) {

	// Generate filter from username and group
	type Search struct {
		Login string
		Group string
	}

	data := Search{
		Login: username,
		Group: group,
	}
	t := template.Must(template.New("filter").Parse(la.settings.UserGroupFilter))
	buf := bytes.NewBufferString("")
	t.Execute(buf, data)
	filter := buf.String()

	attributes := []string{
		"dn",
	}
	if loginField != "" {
		attributes = append(attributes, loginField)
	}

	searchRequest := ldap.NewSearchRequest(
		la.settings.BaseDN, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter, // The filter to apply
		attributes,
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return "", "", err
	}

	if len(sr.Entries) != 1 {
		err := errors.New("LDAP: User `" + username + "' does not exist or too many entries returned")
		return "", "", err
	}

	dn := sr.Entries[0].DN
	if loginField != "" {
		return dn, sr.Entries[0].GetAttributeValue(loginField), nil
	}
	return dn, username, nil
}

// Search connect and search for a username in the ldap, add the entry
func (la *LDAPAuth) Search(username string, group string, loginField string) (string, string, error) {
	l, err := la.connect()
	if err != nil {
		return "", "", err
	}
	defer l.Close()
	return la.search(l, username, group, loginField)
}

// VerifyDN connect, and verify the password of dn identified user
func (la *LDAPAuth) VerifyDN(dn string, password string) error {
	l, err := la.connect()
	if err != nil {
		return err
	}
	defer l.Close()
	return l.Bind(dn, password)
}
