package usecases_test

import (
	"math/rand"
	"testing"

	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestUseCases_IsFeatureGloballyEnabled(t *testing.T) {

	s := testcase.NewSpec(t)
	SetUp(s)
	s.Parallel()

	s.Let(`UseCases`, func(t *testcase.T) interface{} {
		return usecases.NewUseCases(ExampleStorage(t))
	})

	subject := func(t *testcase.T) (bool, error) {
		uc := t.I(`UseCases`).(*usecases.UseCases)
		// TODO: fix this
		return uc.FlagChecker.IsFeatureGloballyEnabled(t.I(`ReleaseFlagName`).(string))
	}

	isEnrolled := func(t *testcase.T) bool {
		enrolled, err := subject(t)
		require.Nil(t, err)
		return enrolled
	}

	s.When(`flag is already configured`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			EnsureFlag(t, ExampleReleaseFlagName(t), t.I(`percentage`).(int))
		})

		s.And(`with global rollout (100%)`, func(s *testcase.Spec) {
			s.Let(`percentage`, func(t *testcase.T) interface{} { return int(100) })

			s.Then(`the feature will be reported to be globally enabled`, func(t *testcase.T) {
				require.True(t, isEnrolled(t))
			})
		})

		s.And(`with less than 100%`, func(s *testcase.Spec) {
			s.Let(`percentage`, func(t *testcase.T) interface{} { return rand.Intn(100) })

			s.Then(`it will report that feature is currently not available globally`, func(t *testcase.T) {
				require.False(t, isEnrolled(t))
			})
		})
	})

}
