package interactors_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	"github.com/stretchr/testify/require"
)

func TestPseudoRandPercentageWithFNV1a64(t *testing.T) {
	subject := interactors.GeneratePseudoRandPercentageWithFNV1a64
	seedSalt := time.Now().Unix()

	t.Run(`it is expected that the result is deterministic`, func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 1000; i++ {
			res1, err1 := subject(strconv.Itoa(i), seedSalt)
			res2, err2 := subject(strconv.Itoa(i), seedSalt)
			require.Nil(t, err1)
			require.Nil(t, err2)
			require.Equal(t, res1, res2)
		}
	})

	t.Run(`it is expected that the values are between 0 and 100`, func(t *testing.T) {
		t.Parallel()

		var minFount, maxFount bool

		for i := 0; i <= 10000; i++ {
			res, err := subject(strconv.Itoa(i), seedSalt)
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
