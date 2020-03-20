package security_test

import (
	"github.com/toggler-io/toggler/domains/security"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestToSHA512Hex(t *testing.T) {
	willBe := func(originalText, expected string) {
		actually, err := security.ToSHA512Hex(originalText)
		require.Nil(t, err)
		require.NotEmpty(t, actually)
		require.Equal(t, expected, actually)
	}

	willBe(`hello`, `9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043`)
	willBe(`what's up`, `87fff07badef9a4853e76abab30a95794615d33e38e4ff6be6d618ebaf1b32bde4d0add8479653d6d2733703f246b9fff4ba4c48fb9488496d5bd1c46a4312e7`)
	willBe(`that should be enough`, `9578f19acbe3647b6f2aa05814dc65bddb6510cdaf149df19f5d8b9ff3c958045f0c659c70efb2839a29321187df879a52e012d54f1cd73f2411b84596ab65c1`)
}
