package rollouts_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/testcase"

	"github.com/stretchr/testify/require"
)

func TestPseudoRandPercentageGenerator_FNV1a64(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := rollouts.PseudoRandPercentageGenerator{}.FNV1a64

	s.Let(`seedSalt`, func(t *testcase.T) interface{} {
		return time.Now().Unix()
	})

	getSeedSalt := func(t *testcase.T) int64 {
		return t.I(`seedSalt`).(int64)
	}

	s.Then(`it is expected that the result is deterministic`, func(t *testcase.T) {
		for i := 0; i < 1000; i++ {
			res1, err1 := subject(strconv.Itoa(i), getSeedSalt(t))
			res2, err2 := subject(strconv.Itoa(i), getSeedSalt(t))
			require.Nil(t, err1)
			require.Nil(t, err2)
			require.Equal(t, res1, res2)
		}
	})

	s.Then(`it is expected that the values are between 0 and 100`, func(t *testcase.T) {
		var minFount, maxFount bool

		for i := 0; i <= 10000; i++ {
			res, err := subject(strconv.Itoa(i), getSeedSalt(t))
			require.Nil(t, err)

			require.True(t, 0 <= res && res <= 100)

			if res == 0 {
				minFount = true
			}

			if res == 100 {
				maxFount = true
			}
		}

		require.True(t, minFount)
		require.True(t, maxFount)
	})
}
