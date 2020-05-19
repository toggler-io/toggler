package testing

import (
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/security"
)

const (
	TokenLetVar     = `testing token`
	tokenTextLetVar = `testing token as text`
	UniqueUserIDLetVar = `UniqUserID`
)

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(TokenLetVar, func(t *testcase.T) interface{} {
			textToken, objectToken := CreateToken(t, ExampleUniqueUserID(t))
			t.Let(tokenTextLetVar, textToken)
			return objectToken
		})

		s.Let(UniqueUserIDLetVar, func(t *testcase.T) interface{} {
			return RandomUniqUserID()
		})
	})
}

func ExampleTextToken(t *testcase.T) string {
	ExampleToken(t)
	return t.I(tokenTextLetVar).(string)
}

func ExampleToken(t *testcase.T) *security.Token {
	return t.I(TokenLetVar).(*security.Token)
}

func CreateToken(t *testcase.T, tokenOwner string) (string, *security.Token) {
	textToken, token, err := ExampleUseCases(t).Issuer.CreateNewToken(GetContext(t), tokenOwner, nil, nil)
	require.Nil(t, err)
	return textToken, token
}


func ExampleUniqueUserID(t *testcase.T) string {
	return t.I(UniqueUserIDLetVar).(string)
}