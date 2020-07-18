package httputils_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

//X-Forwarded-For: 203.0.113.195, 70.41.3.18, 150.172.238.178

func TestGetClientIP(t *testing.T) {
	s := testcase.NewSpec(t)

	subject := func(t *testcase.T) string {
		return httputils.GetClientIP(t.I(`request`).(*http.Request))
	}
	s.Let(`request`, func(t *testcase.T) interface{} {
		return httptest.NewRequest(http.MethodGet, `/`, nil)
	})

	var createRandomIP = func() string {
		return fmt.Sprintf(`%d.%d.%d.%d`,
			rand.Intn(192), rand.Intn(168), rand.Intn(15), rand.Intn(100))
	}

	var expectedRemoteAddress = createRandomIP()

	s.When(`X-REAL-IP is set by Load Balancing service`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.I(`request`).(*http.Request).Header.Set(`X-Real-IP`, expectedRemoteAddress)
		})

		s.Then(`it will return the x-real-ip value`, func(t *testcase.T) {
			require.Equal(t, expectedRemoteAddress, subject(t))
		})
	})

	s.When(`X-Forwarded-For is set by proxy`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.I(`request`).(*http.Request).Header.Set(`X-Forwarded-For`, expectedRemoteAddress)
		})

		s.Then(`it will return the x-forwarded-for value`, func(t *testcase.T) {
			require.Equal(t, expectedRemoteAddress, subject(t))
		})

		s.And(`the request went trough multiple proxy, and all of them appended the previous caller ip by comma"`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				h := t.I(`request`).(*http.Request).Header
				h.Set(`X-Forwarded-For`, fmt.Sprintf(`%s, %s`, h.Get(`X-Forwarded-For`), createRandomIP()))
			})

			s.Then(`it will return only the first x-forwarded-for value`, func(t *testcase.T) {
				require.Equal(t, expectedRemoteAddress, subject(t))
			})
		})
	})

	s.When(`remote-address is set by the http.Server implementation`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.I(`request`).(*http.Request).RemoteAddr = t.I(`remote addr`).(string)
		})

		s.And(`it only includes the HOST`, func(s *testcase.Spec) {
			s.Let(`remote addr`, func(t *testcase.T) interface{} { return expectedRemoteAddress })

			s.Then(`host is retrieved from the request remote addr`, func(t *testcase.T) {
				require.Equal(t, expectedRemoteAddress, subject(t))
			})
		})

		s.And(`it is using the HOST:PORT syntax`, func(s *testcase.Spec) {
			s.Let(`remote addr`, func(t *testcase.T) interface{} {
				return fmt.Sprintf(`%s:%d`, expectedRemoteAddress, 1000+rand.Intn(8000))
			})

			s.Then(`host is retrieved from the request remote addr, but only the host part is kept`, func(t *testcase.T) {
				require.Equal(t, expectedRemoteAddress, subject(t))
			})
		})
	})
}
