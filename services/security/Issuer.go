package security

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/pkg/errors"
	"strings"
	"time"
)

func NewIssuer(s Storage) *Issuer {
	return &Issuer{Storage: s}
}

type Issuer struct {
	Storage
}

func (i *Issuer) CreateNewToken(userUID string, issueAt *time.Time, duration *time.Duration) (*Token, error) {

	if userUID == `` {
		return nil, errors.New(`UserUID cannot be empty`)
	}

	token := &Token{UserUID: userUID}

	if issueAt == nil {
		token.IssuedAt = time.Now().UTC()
	} else {
		token.IssuedAt = *issueAt
	}

	if duration != nil {
		token.Duration = *duration
	}

	tToken, err := i.generateToken()

	if err != nil {
		return nil, err
	}

	token.Token = tToken

	return token, i.Storage.Save(token)

}

func (i *Issuer) RevokeToken(token *Token) error {
	return i.Storage.DeleteByID(token, token.ID)
}

const tokenRawLength = 128

// generateToken returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func (i *Issuer) generateToken() (string, error) {
	b := make([]byte, tokenRawLength)
	_, err := rand.Read(b)
	t := base64.URLEncoding.EncodeToString(b)
	t = strings.TrimRight(t, `=`)
	return t, err
}
