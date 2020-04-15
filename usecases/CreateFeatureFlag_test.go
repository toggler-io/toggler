package usecases_test

import (
	"context"
	. "github.com/toggler-io/toggler/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUseCases_CreateFeatureFlagRolloutStrategyToUseDecisionLogicAPI(t *testing.T) {
	s := testcase.NewSpec(t)
	SetUp(s)
	s.Parallel()

	subject := func(t *testcase.T) error {
		return ExampleUseCases(t).CreateFeatureFlag(context.TODO(), ExampleReleaseFlag(t))
	}

	s.When(`with valid values`, func(s *testcase.Spec) {
		s.Then(`it will be set/persisted`, func(t *testcase.T) {
			require.Nil(t, subject(t))

			require.Equal(t, ExampleReleaseFlag(t), FindStoredExampleReleaseFlagByName(t))
		})
	})

}
