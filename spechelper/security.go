package spechelper

import (
	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/security"
)

const (
	LetVarExampleToken = `example token`
	LetVarTokenText    = `example token as text`
	LetVarUniqueUserID = `UniqUserID`
)

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(LetVarExampleToken, func(t *testcase.T) interface{} {
			textToken, objectToken := CreateToken(t, ExampleUniqueUserID(t))
			t.Set(LetVarTokenText, textToken)
			return objectToken
		})

		s.Let(LetVarUniqueUserID, func(t *testcase.T) interface{} {
			return fixtures.Random.String()
		})
	})
}

func ExampleTextToken(t *testcase.T) string {
	ExampleToken(t)
	return t.I(LetVarTokenText).(string)
}

func ExampleToken(t *testcase.T) *security.Token {
	return t.I(LetVarExampleToken).(*security.Token)
}

func CreateToken(t *testcase.T, tokenOwner string) (string, *security.Token) {
	textToken, token, err := ExampleUseCases(t).Issuer.CreateNewToken(ContextGet(t), tokenOwner, nil, nil)
	require.Nil(t, err)

	storage := StorageGet(t).SecurityToken(ContextGet(t))
	found, err := storage.FindByID(ContextGet(t), &security.Token{}, token.ID)
	require.Nil(t, err)
	require.True(t, found)
	t.Defer(storage.DeleteByID, ContextGet(t), token.ID)

	t.Logf(`%#v - %s`, token, textToken)
	return textToken, token
}

func ExampleUniqueUserID(t *testcase.T) string {
	return t.I(LetVarUniqueUserID).(string)
}
