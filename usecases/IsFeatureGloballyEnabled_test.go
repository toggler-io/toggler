package usecases_test

import (
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestUseCases_IsFeatureGloballyEnabled(t *testing.T) {

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`UseCases`, func(t *testcase.T) interface{} {
		return usecases.NewUseCases(t.I(`TestStorage`).(*TestStorage))
	})

	subject := func(t *testcase.T) (bool, error) {
		uc := t.I(`UseCases`).(*usecases.UseCases)
		return uc.IsFeatureGloballyEnabled(t.I(`FeatureName`).(string))
	}

	isEnrolled := func(t *testcase.T) bool {
		enrolled, err := subject(t)
		require.Nil(t, err)
		return enrolled
	}

	s.When(`flag is already configured`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, t.I(`UseCases`).(*usecases.UseCases).
				UpdateFeatureFlagRolloutPercentage(t.I(`FeatureName`).(string),
					t.I(`percentage`).(int)))
		})

		s.And(`with global rollout (100%)`, func(s *testcase.Spec) {
			s.Let(`percentage`, func(t *testcase.T) interface{} { return int(100) })

			s.Then(`the feature will be reportad to be globally enabled`, func(t *testcase.T) {
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