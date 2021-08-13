package release_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

var (
	_ release.RolloutPlan = release.RolloutDecisionByPercentage{}
	_ release.RolloutPlan = release.RolloutDecisionByAPI{}
	_ release.RolloutPlan = release.RolloutDecisionAND{}
	_ release.RolloutPlan = release.RolloutDecisionOR{}
)

//--------------------------------------------------------------------------------------------------------------------//

func TestRolloutDecisionByPercentage(t *testing.T) {
	s := sh.NewSpec(t)

	var rolloutDefinition = func(t *testcase.T) release.RolloutDecisionByPercentage {
		r := release.NewRolloutDecisionByPercentage()
		r.PseudoRandPercentageAlgorithm = `func`
		r.PseudoRandPercentageFunc = func(id string, seedSalt int64) (int, error) {
			err, _ := t.I(`pseudo rand error`).(error)
			percentage, _ := t.I(`pseudo rand percentage`).(int)
			return percentage, err
		}
		r.Percentage = t.I(`percentage`).(int)
		return r
	}

	s.Let(`pseudo rand error`, func(t *testcase.T) interface{} { return nil }) // no error is the default case

	s.Describe(`IsParticipating`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) (bool, error) {
			return rolloutDefinition(t).IsParticipating(sh.ContextGet(t), sh.ExampleExternalPilotID(t))
		}

		var onSuccess = func(t *testcase.T) bool {
			ok, err := subject(t)
			require.Nil(t, err)
			return ok
		}

		var andRandGeneratorEncounterAnError = func(s *testcase.Spec) {
			s.And(`rand generator encounter an error`, func(s *testcase.Spec) {
				const expectedError frameless.Error = `boom`
				s.LetValue(`pseudo rand error`, expectedError)

				s.Then(`it will propagate back the error`, func(t *testcase.T) {
					_, actualErr := subject(t)
					require.Equal(t, expectedError, actualErr)
				})
			})
		}

		s.When(`configured percentage`, func(s *testcase.Spec) {
			s.Context(`it is 0`, func(s *testcase.Spec) {
				s.Let(`percentage`, func(t *testcase.T) interface{} { return int(0) })

				s.And(`pseudo rand percentage for the id`, func(s *testcase.Spec) {
					s.And(`it is 0 as well`, func(s *testcase.Spec) {
						s.Let(`pseudo rand percentage`, func(t *testcase.T) interface{} { return int(0) })

						s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
							require.False(t, onSuccess(t))
						})
					})

					s.And(`it is greater than 0`, func(s *testcase.Spec) {
						s.Let(`pseudo rand percentage`, func(t *testcase.T) interface{} { return int(42) })

						s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
							require.False(t, onSuccess(t))
						})
					})
				})
			})

			s.Context(`it is greater than 0`, func(s *testcase.Spec) {
				s.Let(`percentage`, func(t *testcase.T) interface{} { return fixtures.Random.IntBetween(1, 99) })

				s.And(`the next pseudo rand int is less or eq with the rolloutDefinition percentage`, func(s *testcase.Spec) {
					s.Let(`pseudo rand percentage`, func(t *testcase.T) interface{} {
						return fixtures.Random.IntBetween(0, t.I(`percentage`).(int))
					})

					s.Then(`it will enroll the pilot for the feature`, func(t *testcase.T) {
						require.True(t, onSuccess(t))
					})

					andRandGeneratorEncounterAnError(s)
				})

				s.And(`the next pseudo rand int is is greater than the rolloutDefinition percentage`, func(s *testcase.Spec) {
					s.Let(`pseudo rand percentage`, func(t *testcase.T) interface{} {
						return fixtures.Random.IntBetween(t.I(`percentage`).(int)+1, 100)
					})

					s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
						require.False(t, onSuccess(t))
					})

					andRandGeneratorEncounterAnError(s)
				})
			})

			s.Context(`it is 100% or in other words it is set to be globally enabled`, func(s *testcase.Spec) {
				s.Let(`percentage`, func(t *testcase.T) interface{} { return int(100) })

				s.And(`basically regardless the pseudo random percentage`, func(s *testcase.Spec) {
					s.Let(`pseudo rand percentage`, func(t *testcase.T) interface{} {
						return fixtures.Random.IntBetween(0, 100)
					})

					s.Then(`it will always be enrolled`, func(t *testcase.T) {
						require.True(t, onSuccess(t))
					})

					andRandGeneratorEncounterAnError(s)
				})
			})
		})
	})
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRolloutDecisionByAPI(t *testing.T) {
	s := sh.NewSpec(t)

	var rolloutDefinition = func(t *testcase.T) release.RolloutDecisionByAPI {
		r := release.NewRolloutDecisionByAPIDeprecated()
		r.URL, _ = t.I(`url`).(*url.URL)
		return r
	}

	s.Describe(`IsParticipating`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) (bool, error) {
			return rolloutDefinition(t).IsParticipating(sh.ContextGet(t), sh.ExampleExternalPilotID(t))
		}

		const testServerLetVar = `httptest.NewServer`

		s.Let(`url`, func(t *testcase.T) interface{} {
			server := t.I(testServerLetVar).(*httptest.Server)
			u, err := url.Parse(server.URL)
			require.Nil(t, err)
			return u
		})

		s.Let(testServerLetVar, func(t *testcase.T) interface{} {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				require.Equal(t, sh.ExampleExternalPilotID(t), q.Get(`pilot-external-id`))
				w.WriteHeader(t.I(`replyCode`).(int))
			}))
			t.Defer(s.Close)
			return s
		})

		s.And(`the remote encountered an unexpected error`, func(s *testcase.Spec) {
			s.Let(`replyCode`, func(t *testcase.T) interface{} { return fixtures.Random.IntBetween(500, 599) })

			s.Then(`pilot enrollment check reported as failure`, func(t *testcase.T) {
				_, err := subject(t)
				require.Error(t, err)
			})
		})

		s.And(`the remote reject the pilot enrollment`, func(s *testcase.Spec) {
			s.Let(`replyCode`, func(t *testcase.T) interface{} { return fixtures.Random.IntBetween(400, 499) })

			s.Then(`the pilot is not enrolled for the feature`, func(t *testcase.T) {
				enabled, err := subject(t)
				require.Nil(t, err)
				require.False(t, enabled)
			})
		})

		s.And(`the the remote accept the pilot enrollment`, func(s *testcase.Spec) {
			s.Let(`replyCode`, func(t *testcase.T) interface{} { return fixtures.Random.IntBetween(200, 299) })

			s.Then(`the pilot is not enrolled for the feature`, func(t *testcase.T) {
				enabled, err := subject(t)
				require.Nil(t, err)
				require.True(t, enabled)
			})
		})
	})

	s.Describe(`Validate`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) error {
			return rolloutDefinition(t).Validate()
		}

		s.When(`url is valid`, func(s *testcase.Spec) {
			s.Let(`url`, func(t *testcase.T) interface{} {
				u, err := url.Parse(`https://example.com`)
				require.Nil(t, err)
				return u
			})

			s.Then(`it is valid`, func(t *testcase.T) {
				require.Nil(t, subject(t))
			})
		})

		s.When(`url is invalid`, func(s *testcase.Spec) {
			s.Let(`url`, func(t *testcase.T) interface{} {
				return &url.URL{}
			})

			s.Then(`it yields error`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})

		s.When(`url is nil`, func(s *testcase.Spec) {
			s.Let(`url`, func(t *testcase.T) interface{} {
				return nil
			})

			s.Then(`it yields error`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})
	})
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRolloutDecisionByIPAddress(t *testing.T) {
	t.Skip()
	//s :=testcase.NewSpec(t)
	//subject := func(t *testcase.T) error {
	//	rm := t.I(`RolloutManager`).(*release.RolloutManager)
	//	return rm.AllowIPAddrForFlag(
	//		GetContext(t),
	//		getReleaseFlag(t).ID,
	//		t.I(`ip-addr`).(string),
	//	)
	//}
	//
	//s.Before(func(t *testcase.T) { t.Skip(`WIP`) })
	//
	//s.Before(func(t *testcase.T) {
	//	t.Log(`the flag has 0 as release percentage`)
	//	getReleaseFlag(t).Rollout.Strategy.Percentage = 0
	//	t.Log(`the flag is saved in the storage`)
	//	require.Nil(t, StorageGet(t).Create(GetContext(t), getReleaseFlag(t)))
	//})
	//
	//s.Let(`ip-addr`, func(t *testcase.T) interface{} {
	//	return fmt.Sprintf(`192.168.1.%d`, rand.Intn(255))
	//})
	//
	//s.Then(`upon calling it, should result in no error`, func(t *testcase.T) {
	//	require.Nil(t, subject(t))
	//})
	//
	//s.When(`ip address value is an invalid value`, func(s *testcase.Spec) {
	//	s.Let(`ip-addr`, func(t *testcase.T) interface{} {
	//		return `invalid-value`
	//	})
	//
	//	s.Then(`it should report error about it`, func(t *testcase.T) {
	//		t.Skip(`extend test coverage with IPv4 and IPv6 happy paths before this is acceptable`)
	//	})
	//})
	//
	//s.Describe(`relation with GetAllReleaseFlagStatesOfThePilot`, func(s *testcase.Spec) {
	//	releaseFlagState := func(t *testcase.T) bool {
	//		fc := release.NewFlagChecker(StorageGet(t))
	//		ctx := context.WithValue(GetContext(t), release.CtxPilotIpAddr, t.I(`ip-addr`).(string))
	//		states, err := fc.GetAllReleaseFlagStatesOfThePilot(ctx, ExampleReleaseFlagName(t), *ExampleDeploymentEnvironment(t), ExampleExternalPilotID(t))
	//		require.Nil(t, err)
	//		return states[getReleaseFlag(t).Name]
	//	}
	//
	//	s.When(`the ip allow value is saved`, func(s *testcase.Spec) {
	//		s.Before(func(t *testcase.T) { require.Nil(t, subject(t)) })
	//
	//		s.Then(`the flag will be stated as enabled`, func(t *testcase.T) {
	//			require.True(t, releaseFlagState(t))
	//		})
	//	})
	//
	//	s.When(`the ip allow value is not persisted`, func(s *testcase.Spec) {
	//		// nothing to do here, as implicitly this is achieved by it
	//
	//		s.Then(`the flag will be stated as disabled`, func(t *testcase.T) {
	//			require.False(t, releaseFlagState(t))
	//		})
	//	})
	//})
}

//--------------------------------------------------------------------------------------------------------------------//

func TestPseudoRandPercentageGenerator_FNV1a64(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := release.PseudoRandPercentageAlgorithms{}.FNV1a64

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

	s.Then(`distributes evenly`, func(t *testcase.T) {
		var (
			seedSalt             = int64(fixtures.Random.Int())
			samplingCount        int
			maxPercentage        = fixtures.Random.IntBetween(30, 70)
			tolerationPercentage int
		)
		if testing.Short() {
			tolerationPercentage = 5
			samplingCount = 1000
		} else {
			tolerationPercentage = 3
			samplingCount = 10000
		}
		require.True(t, maxPercentage+tolerationPercentage <= 100)

		var randomExternalPilotID = func() string {
			blocks := make([]string, 0, 6)
			for i := 0; i < 6; i++ {
				number := fmt.Sprintf("%02x", fixtures.Random.IntN(255))
				blocks = append(blocks, number)
			}
			return strings.Join(blocks, ":")
		}

		extIDS := make([]string, 0, samplingCount)
		for i := 0; i < samplingCount; i++ {
			extIDS = append(extIDS, randomExternalPilotID())
		}

		/* start E2E test */

		var enrolled, rejected int
		for _, extID := range extIDS {
			p, err := subject(extID, seedSalt)
			require.Nil(t, err)

			if p <= maxPercentage {
				enrolled++
			} else {
				rejected++
			}
		}

		require.True(t, enrolled > 0, `no one enrolled? fishy`)

		t.Logf(`a little toleration is still accepted, as long in generally it is within the range (+%d%%)`, tolerationPercentage)
		maximumAcceptedEnrollmentPercentage := maxPercentage + tolerationPercentage
		minimumAcceptedEnrollmentPercentage := maxPercentage - tolerationPercentage

		t.Logf(`so the total percentage in this test that fulfil the requirements is %d%%`, maximumAcceptedEnrollmentPercentage)
		testRunResultPercentage := int(float32(enrolled) / float32(enrolled+rejected) * 100)

		t.Logf(`and the actual percentage is %d%%`, testRunResultPercentage)
		require.True(t, testRunResultPercentage <= maximumAcceptedEnrollmentPercentage)
		require.True(t, minimumAcceptedEnrollmentPercentage <= testRunResultPercentage)
	})

}

//--------------------------------------------------------------------------------------------------------------------//

func TestRollout(t *testing.T) {
	s := sh.NewSpec(t)

	var rollout = func(t *testcase.T) *release.Rollout { return t.I(`rollout`).(*release.Rollout) }
	s.Let(`rollout`, func(t *testcase.T) interface{} {
		plan, _ := t.I(`plan`).(release.RolloutPlan)
		return &release.Rollout{
			FlagID:        sh.ExampleReleaseFlag(t).ID,
			EnvironmentID: sh.ExampleDeploymentEnvironment(t).ID,
			Plan:          plan,
		}
	})

	s.Describe(`#Validate`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) error {
			return rollout(t).Validate()
		}

		s.Let(`plan`, func(t *testcase.T) interface{} {
			return release.NewRolloutDecisionByPercentage()
		})

		s.When(`env id is not set`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				rollout(t).EnvironmentID = ``
			})

			s.Then(`it will yield error about it`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})

		s.When(`flag id is not set`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				rollout(t).FlagID = ``
			})

			s.Then(`it will yield error about it`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})

		s.When(`plan invalid`, func(s *testcase.Spec) {
			s.Let(`plan`, func(t *testcase.T) interface{} {
				p := release.NewRolloutDecisionByPercentage()
				p.Percentage = 120
				return p
			})

			s.Then(`it will yield error about it`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})

		s.When(`plan is not provided`, func(s *testcase.Spec) {
			s.Let(`plan`, func(t *testcase.T) interface{} { return nil })

			s.Then(`it will yield error about it`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})
	})

	s.Describe(`JSON`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			rIn := rollout(t)

			bs, err := json.Marshal(rIn)
			require.Nil(t, err)

			var rOut release.Rollout
			require.Nil(t, json.Unmarshal(bs, &rOut))

			require.Equal(t, *rIn, rOut)
		}

		s.When(`plan is`, func(s *testcase.Spec) {
			s.Context(`RolloutDecisionByPercentage`, func(s *testcase.Spec) {
				s.Let(`plan`, func(t *testcase.T) interface{} {
					plan := release.NewRolloutDecisionByPercentage()
					plan.Percentage = fixtures.Random.IntBetween(50, 70)
					return plan
				})

				s.Then(`marshal and unmarshal back and forth`, func(t *testcase.T) { subject(t) })
			})

			s.Context(`RolloutDecisionByAPI`, func(s *testcase.Spec) {
				s.Let(`plan`, func(t *testcase.T) interface{} {
					plan := release.NewRolloutDecisionByAPIDeprecated()
					u, err := url.Parse(`https://example.com`)
					require.Nil(t, err)
					plan.URL = u
					return plan
				})

				s.Then(`marshal and unmarshal back and forth`, func(t *testcase.T) { subject(t) })
			})

			s.Context(`RolloutDecisionAND`, func(s *testcase.Spec) {
				s.Let(`plan`, func(t *testcase.T) interface{} {
					return release.RolloutDecisionAND{
						Left:  release.NewRolloutDecisionByPercentage(),
						Right: release.NewRolloutDecisionByPercentage(),
					}
				})

				s.Then(`marshal and unmarshal back and forth`, func(t *testcase.T) { subject(t) })
			})

			s.Context(`RolloutDecisionOR`, func(s *testcase.Spec) {
				s.Let(`plan`, func(t *testcase.T) interface{} {
					return release.RolloutDecisionOR{
						Left:  release.NewRolloutDecisionByPercentage(),
						Right: release.NewRolloutDecisionByPercentage(),
					}
				})

				s.Then(`marshal and unmarshal back and forth`, func(t *testcase.T) { subject(t) })
			})

			s.Context(`RolloutDecisionNOT`, func(s *testcase.Spec) {
				s.Let(`plan`, func(t *testcase.T) interface{} {
					return release.RolloutDecisionNOT{Definition: release.NewRolloutDecisionByPercentage()}
				})

				s.Then(`marshal and unmarshal back and forth`, func(t *testcase.T) { subject(t) })
			})
		})
	})

}

func TestRollout_MarshalJSON_e2e(t *testing.T) {
	tc := testcase.NewT(t, testcase.NewSpec(t))

	expected := release.Rollout{
		ID:            tc.Random.String(),
		FlagID:        tc.Random.String(),
		EnvironmentID: tc.Random.String(),
		Plan: release.RolloutDecisionAND{
			Left: release.RolloutDecisionOR{
				Left: release.RolloutDecisionNOT{
					Definition: release.RolloutDecisionByGlobal{
						State: tc.Random.Bool(),
					},
				},
				Right: release.NewRolloutDecisionByAPI(&url.URL{
					Scheme: "https",
					User:   url.UserPassword("foo", "bar"),
					Host:   "www.toggler.io",
					Path:   "/ping",
				}),
			},
			Right: release.RolloutDecisionByPercentage{
				PseudoRandPercentageAlgorithm: "FNV1a64",
				PseudoRandPercentageFunc:      nil,
				Seed:                          int64(tc.Random.Int()),
				Percentage:                    tc.Random.IntBetween(0, 100),
			},
		},
	}

	bs, err := json.MarshalIndent(expected, "", "\t")
	require.Nil(t, err)
	t.Log(string(bs))
	var actual release.Rollout
	require.Nil(t, json.Unmarshal(bs, &actual))
	require.Equal(t, expected, actual)
}
