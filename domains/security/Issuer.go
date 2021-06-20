package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

func NewIssuer(s Storage) *Issuer {
	return &Issuer{Storage: s}
}

type Issuer struct {
	Storage
}

func (i *Issuer) CreateNewToken(ctx context.Context, ownerUID string, issueAt *time.Time, duration *time.Duration) (string, *Token, error) {

	if ownerUID == `` {
		return "", nil, errors.New(`OwnerUID cannot be empty`)
	}

	token := &Token{OwnerUID: ownerUID}

	if issueAt == nil {
		token.IssuedAt = time.Now().UTC()
	} else {
		token.IssuedAt = *issueAt
	}

	if duration != nil {
		token.Duration = *duration
	}

	textToken, err := i.generateToken()

	if err != nil {
		return "", nil, err
	}

	token.SHA512, err = ToSHA512Hex(textToken)

	if err != nil {
		return textToken, token, err
	}

	return textToken, token, i.Storage.SecurityToken(ctx).Create(ctx, token)

}

func (i *Issuer) RevokeToken(ctx context.Context, token *Token) error {
	return i.Storage.SecurityToken(ctx).DeleteByID(ctx, token.ID)
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
