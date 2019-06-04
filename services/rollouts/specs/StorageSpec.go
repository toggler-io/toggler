package specs

import (
	testing2 "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/frameless/resources/specs"
)

type StorageSpec struct {
	Storage rollouts.Storage
}

func (spec *StorageSpec) Test(t *testing.T) {

	entityTypes := []interface{}{
		rollouts.FeatureFlag{},
		rollouts.Pilot{},
	}

	ff := testing2.NewFixtureFactory()

	for _, entityType := range entityTypes {
		specs.TestMinimumRequirements(t, spec.Storage, entityType, ff)
		specs.TestUpdate(t, spec.Storage, entityType, ff)
		specs.TestFindAll(t, spec.Storage, entityType, ff)
	}

	FlagFinderSpec{Subject: spec.Storage}.Test(t)
	PilotFinderSpec{Subject: spec.Storage}.Test(t)

	s := testcase.NewSpec(t)

	s.Describe(`flag name uniq across storage`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) error {
			return spec.Storage.Save(t.I(`flag`).(*rollouts.FeatureFlag))
		}

		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Storage.Truncate(rollouts.FeatureFlag{}))
		})

		s.Let(`flag`, func(t *testcase.T) interface{} {
			return &rollouts.FeatureFlag{
				Name: `my-uniq-flag-name`,
			}
		})

		s.When(`flag already stored`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { require.Nil(t, subject(t)) })

			s.Then(`saving again will create error`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})
	})

}
