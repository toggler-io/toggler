package release_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestFlag(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`flag`, func(t *testcase.T) interface{} {
		return sh.Create(release.Flag{})
	})

	s.Describe(`Validate`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) error {
			return t.I(`flag`).(*release.Flag).Validate()
		}

		s.When(`values are correct`, func(s *testcase.Spec) {
			s.Then(`it should be ok`, func(t *testcase.T) {
				require.Nil(t, subject(t))
			})
		})

		s.When(`name is empty`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { sh.GetReleaseFlag(t, `flag`).Name = `` })

			s.Then(`error reported`, func(t *testcase.T) {
				require.Equal(t, release.ErrNameIsEmpty, subject(t))
			})
		})
	})
}
