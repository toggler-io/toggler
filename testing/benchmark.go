package testing

import (
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
	"testing"
)

func CreateEntities(count int , f specs.FixtureFactory, T interface{}) []interface{} {
	var es []interface{}
	for i := 0; i < count; i++ {
		es = append(es, f.Create(T))
	}
	return es
}

func SaveEntities(b *testing.B, r resources.Creator, f specs.FixtureFactory, es ...interface{}) {
	for _, e := range es {
		require.Nil(b, r.Create(f.Context(), e))
	}
}

func TruncateEntities(b *testing.B, t resources.Truncater, f specs.FixtureFactory, T interface{}) {
	require.Nil(b, t.Truncate(f.Context(), T))
}
