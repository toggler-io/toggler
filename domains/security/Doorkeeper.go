package security

import (
	"context"
)

func NewDoorkeeper(s Storage) *Doorkeeper {
	return &Doorkeeper{Storage: s}
}

type Doorkeeper struct {
	Storage
}

func (dk *Doorkeeper) VerifyTextToken(ctx context.Context, textToken string) (bool, error) {
	sha512hex, err := ToSHA512Hex(textToken)

	if err != nil {
		return false, err
	}

	token, err := dk.Storage.SecurityToken(ctx).FindTokenBySHA512Hex(ctx, sha512hex)

	if token == nil {
		return false, nil
	}

	return token.IsValid(), err
}
