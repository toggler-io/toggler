package release

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Rollout struct {
	ID string `ext:"ID"`
	// FlagID is the release flag id to which the rolloutBase belongs
	FlagID string `json:"flag_id"`
	// EnvironmentID is the deployment environment id
	DeploymentEnvironmentID string `json:"env_id"`
	// Plan holds the composited rule set about the pilot participation decision logic.
	Plan RolloutDefinition `json:"plan,omitempty"`
}

func (r Rollout) Validate() error {
	if r.FlagID == `` {
		return ErrMissingFlag
	}

	if r.DeploymentEnvironmentID == `` {
		return ErrMissingEnv
	}

	if r.Plan == nil {
		return ErrMissingRolloutPlan
	}

	return r.Plan.Validate()
}

// RolloutDefinition is the common interface to all rollout type.
// Rollout expects to determines the behavior of the rollout process.
// the actual behavior implementation is with the RolloutManager,
// but the configuration data is located here
type RolloutDefinition interface {
	IsParticipating(ctx context.Context, pilotExternalID string) (bool, error)
	Validate() error
}

//--------------------------------------------------------------------------------------------------------------------//

// TODO: add proper coverage
func NewRolloutDecisionGlobal() RolloutDecisionByGlobal {
	return RolloutDecisionByGlobal{}
}

type RolloutDecisionByGlobal struct {
	State bool
}

func (r RolloutDecisionByGlobal) IsParticipating(ctx context.Context, pilotExternalID string) (bool, error) {
	return r.State, nil
}

func (r RolloutDecisionByGlobal) Validate() error {
	return nil
}

//--------------------------------------------------------------------------------------------------------------------//

func NewRolloutDecisionByPercentage() RolloutDecisionByPercentage {
	return RolloutDecisionByPercentage{
		PseudoRandPercentageAlgorithm: "FNV1a64",
		PseudoRandPercentageFunc:      nil,
		Seed:                          time.Now().Unix(),
		Percentage:                    0,
	}
}

type RolloutDecisionByPercentage struct {
	// PseudoRandPercentageAlgorithm specifies the algorithm that is expected to be used in the percentage calculation.
	// Currently it is only supports FNV1a64 and "func"
	PseudoRandPercentageAlgorithm string
	// PseudoRandPercentageFunc is a dependency that can be used if the Algorithm is not defined or defined to func.
	// Ideal for testing.
	PseudoRandPercentageFunc func(id string, seedSalt int64) (int, error)
	// Seed allows you to configure the randomness for the percentage based pilot enrollment selection.
	// This value could have been neglected by using the flag name as random seed,
	// but that would reduce the flexibility for edge cases where you want
	// to use a similar pilot group as a successful flag rolloutBase before.
	Seed int64
	// Percentage allows you to define how many of your user base should be enrolled pseudo randomly.
	Percentage int
}

func (s RolloutDecisionByPercentage) IsParticipating(ctx context.Context, pilotExternalID string) (bool, error) {
	if s.Percentage == 0 {
		return false, nil
	}

	diceRollResultPercentage, err := s.pseudoRandPercentage(pilotExternalID)
	if err != nil {
		return false, err
	}

	return diceRollResultPercentage <= s.Percentage, nil
}

func (s RolloutDecisionByPercentage) Validate() error {
	if s.Percentage < 0 || 100 < s.Percentage {
		return ErrInvalidPercentage
	}

	return nil
}

func (s RolloutDecisionByPercentage) pseudoRandPercentage(pilotExternalID string) (int, error) {
	switch strings.ToLower(s.PseudoRandPercentageAlgorithm) {
	case `func`:
		return s.PseudoRandPercentageFunc(pilotExternalID, s.Seed)
	case `fnv1a64`:
		return PseudoRandPercentageAlgorithms{}.FNV1a64(pilotExternalID, s.Seed)
	default:
		return 0, fmt.Errorf(`unknown pseudo rand percentage algorithm: %s`, s.PseudoRandPercentageAlgorithm)
	}
}

// PseudoRandPercentageFunc implements pseudo random percentage calculations with different algorithms.
// This is mainly used for pilot enrollments when percentage strategy is used for rolloutBase.
type PseudoRandPercentageAlgorithms struct{}

// FNV1a64 implements pseudo random percentage calculation with FNV-1a64.
func (g PseudoRandPercentageAlgorithms) FNV1a64(id string, seedSalt int64) (int, error) {
	h := fnv.New64a()

	if _, err := h.Write([]byte(id)); err != nil {
		return 0, err
	}

	seed := int64(h.Sum64()) + seedSalt
	source := rand.NewSource(seed)
	return rand.New(source).Intn(101), nil
}

//--------------------------------------------------------------------------------------------------------------------//

// TODO: implement this as a feature
//type RolloutByJavaScript struct {
//	rolloutBase
//	Script []byte
//}
//
//type RolloutByLua struct {
//	rolloutBase
//	Script []byte
//}

//--------------------------------------------------------------------------------------------------------------------//

func NewRolloutDecisionByAPI() RolloutDecisionByAPI {
	return RolloutDecisionByAPI{
		HTTPClient: http.Client{Timeout: 30 * time.Second},
		URL:        nil,
	}
}

type RolloutDecisionByAPI struct {
	HTTPClient http.Client
	// URL allow you to do rolloutBase based on custom domain needs such as target groups,
	// which decision logic is available trough an API endpoint call
	URL *url.URL
}

func (s RolloutDecisionByAPI) IsParticipating(ctx context.Context, pilotExternalID string) (bool, error) {
	req, err := http.NewRequest(`GET`, s.URL.String(), nil)
	if err != nil {
		return false, err
	}
	req = req.WithContext(ctx)

	query := req.URL.Query()
	query.Set(`pilot-external-id`, pilotExternalID)
	req.URL.RawQuery = query.Encode()

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}

	code := resp.StatusCode

	if 500 <= code && code < 600 {
		defer resp.Body.Close()
		if bs, err := ioutil.ReadAll(resp.Body); err != nil {
			return false, err
		} else {
			return false, fmt.Errorf(string(bs))
		}
	}

	return 200 <= code && code < 300, nil
}

func (s RolloutDecisionByAPI) Validate() error {
	if s.URL == nil {
		return ErrInvalidRequestURL
	}

	_, err := url.ParseRequestURI(s.URL.String())
	if err != nil {
		return ErrInvalidRequestURL
	}

	if s.URL.Scheme == `` {
		return ErrInvalidRequestURL
	}

	if s.URL.Hostname() == `` {
		return ErrInvalidRequestURL
	}

	return nil
}

//--------------------------------------------------------------------------------------------------------------------//

type RolloutDecisionAND struct {
	Left  RolloutDefinition
	Right RolloutDefinition
}

// TODO:SPEC
func (r RolloutDecisionAND) IsParticipating(ctx context.Context, pilotExternalID string) (bool, error) {
	lp, err := r.Left.IsParticipating(ctx, pilotExternalID)
	if err != nil {
		return false, err
	}
	rp, err := r.Right.IsParticipating(ctx, pilotExternalID)
	if err != nil {
		return false, err
	}
	return lp && rp, nil
}

// TODO:SPEC
func (r RolloutDecisionAND) Validate() error {
	if r.Left == nil {
		return fmt.Errorf(`left rollout decision node is missing in AND`)
	}
	if err := r.Left.Validate(); err != nil {
		return err
	}
	if r.Right == nil {
		return fmt.Errorf(`right rollout decision node is missing in AND`)
	}
	if err := r.Right.Validate(); err != nil {
		return err
	}
	return nil
}

//--------------------------------------------------------------------------------------------------------------------//

type RolloutDecisionOR struct {
	Left  RolloutDefinition
	Right RolloutDefinition
}

// TODO:SPEC
func (r RolloutDecisionOR) IsParticipating(ctx context.Context, pilotExternalID string) (bool, error) {
	lp, err := r.Left.IsParticipating(ctx, pilotExternalID)
	if err != nil {
		return false, err
	}
	rp, err := r.Right.IsParticipating(ctx, pilotExternalID)
	if err != nil {
		return false, err
	}
	return lp || rp, nil
}

// TODO:SPEC
func (r RolloutDecisionOR) Validate() error {
	if r.Left == nil {
		return fmt.Errorf(`left rollout decision node is missing in OR`)
	}
	if err := r.Left.Validate(); err != nil {
		return err
	}
	if r.Right == nil {
		return fmt.Errorf(`right rollout decision node is missing in OR`)
	}
	if err := r.Right.Validate(); err != nil {
		return err
	}
	return nil
}

//--------------------------------------------------------------------------------------------------------------------//

type RolloutDecisionNOT struct {
	Definition RolloutDefinition
}

// TODO:SPEC
func (r RolloutDecisionNOT) IsParticipating(ctx context.Context, pilotExternalID string) (bool, error) {
	p, err := r.Definition.IsParticipating(ctx, pilotExternalID)
	if err != nil {
		return false, err
	}
	return !p, nil
}

// TODO:SPEC
func (r RolloutDecisionNOT) Validate() error {
	if r.Definition == nil {
		return fmt.Errorf(`rollout decesion logic missing for not`)
	}
	return r.Definition.Validate()
}

/*

  (percentage is 50 AND platform is android) OR ip is "192.0.2.1"

	*translates into*

  OR - AND - percentage is 50
           - platform is android
     - ip is "192.0.2.1"

*/

//--------------------------------------------------------------------------------------------------------------------//
//---------------------------------------------------- MARSHALING ----------------------------------------------------//
//--------------------------------------------------------------------------------------------------------------------//

type RolloutDefinitionView struct {
	Definition RolloutDefinition
}

func (view RolloutDefinitionView) MarshalJSON() ([]byte, error) {
	plan, err := view.MarshalMapping(view.Definition)
	if err != nil {
		return nil, err
	}
	return json.Marshal(plan)
}

func (view *RolloutDefinitionView) UnmarshalJSON(bytes []byte) error {
	mapping := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &mapping); err != nil {
		return err
	}

	plan, err := view.UnmarshalMapping(mapping)
	if err != nil {
		return err
	}

	if plan != nil {
		if err := plan.Validate(); err != nil {
			return err
		}
	}

	view.Definition = plan
	return nil
}

func (view RolloutDefinitionView) MarshalMapping(this RolloutDefinition) (map[string]interface{}, error) {
	var m = make(map[string]interface{})
	switch d := this.(type) {
	case RolloutDecisionByGlobal:
		m[`type`] = `global`
		m[`state`] = d.State
		return m, nil

	case RolloutDecisionByPercentage:
		m[`type`] = `percentage`
		m[`percentage`] = d.Percentage
		m[`algorithm`] = d.PseudoRandPercentageAlgorithm
		m[`seed`] = d.Seed
		return m, nil

	case RolloutDecisionByAPI:
		m[`type`] = `api`
		if d.URL == nil {
			return m, fmt.Errorf(`missing url for RolloutDecisionByAPI json marshaling`)
		}
		m[`url`] = d.URL.String()
		return m, nil

	case RolloutDecisionAND:
		m[`type`] = `and`
		var err error
		m[`left`], err = view.MarshalMapping(d.Left)
		if err != nil {
			return m, err
		}
		m[`right`], err = view.MarshalMapping(d.Right)
		if err != nil {
			return m, err
		}
		return m, nil

	case RolloutDecisionOR:
		m[`type`] = `or`
		var err error
		m[`left`], err = view.MarshalMapping(d.Left)
		if err != nil {
			return m, err
		}
		m[`right`], err = view.MarshalMapping(d.Right)
		if err != nil {
			return m, err
		}
		return m, nil

	case RolloutDecisionNOT:
		m[`type`] = `not`
		var err error
		m[`def`], err = view.MarshalMapping(d.Definition)
		if err != nil {
			return m, err
		}
		return m, nil

	case nil:
		return nil, nil

	default:
		return nil, fmt.Errorf(`unknown rollout definition struct: %T`, this)
	}
}

func (view RolloutDefinitionView) UnmarshalMapping(this map[string]interface{}) (_ RolloutDefinition, rErr error) {
	defer func() {
		if r := recover(); r != nil {
			rErr = fmt.Errorf(`%v`, r)
		}
	}()

	if this == nil {
		return nil, nil
	}

	switch strings.ToLower(this[`type`].(string)) {
	case `global`:
		d := NewRolloutDecisionGlobal()
		if v, ok := this[`state`]; ok {
			d.State = v.(bool)
		}
		return d, nil

	case `percentage`:
		d := NewRolloutDecisionByPercentage()
		if v, ok := this[`percentage`]; ok {
			d.Percentage = int(v.(float64))
		}

		if v, ok := this[`algorithm`]; ok {
			d.PseudoRandPercentageAlgorithm = v.(string)
		}

		if v, ok := this[`seed`]; ok {
			d.Seed = int64(v.(float64))
		}

		return d, nil

	case `api`:
		d := NewRolloutDecisionByAPI()
		raw, ok := this[`url`]
		if !ok {
			return d, fmt.Errorf(`missing url for rollout decision by API`)
		}
		u, err := url.Parse(raw.(string))
		if err != nil {
			return d, err
		}
		d.URL = u
		return d, nil

	case `and`:
		var d RolloutDecisionAND

		l, err := view.UnmarshalMapping(this[`left`].(map[string]interface{}))
		if err != nil {
			return d, err
		}
		d.Left = l

		r, err := view.UnmarshalMapping(this[`right`].(map[string]interface{}))
		if err != nil {
			return d, err
		}
		d.Right = r

		return d, nil

	case `or`:
		var d RolloutDecisionOR

		l, err := view.UnmarshalMapping(this[`left`].(map[string]interface{}))
		if err != nil {
			return d, err
		}
		d.Left = l

		r, err := view.UnmarshalMapping(this[`right`].(map[string]interface{}))
		if err != nil {
			return d, err
		}
		d.Right = r

		return d, nil

	case `not`:
		var d RolloutDecisionNOT

		def, err := view.UnmarshalMapping(this[`def`].(map[string]interface{}))
		if err != nil {
			return d, err
		}
		d.Definition = def

		return d, nil

	default:
		return nil, fmt.Errorf(`unknown rollout definition type: %s`, this[`type`])
	}
}

type rolloutView struct {
	ID string `ext:"ID" json:"id"`
	// FlagID is the release flag id to which the rolloutBase belongs
	FlagID string `json:"flag_id"`
	// EnvironmentID is the deployment environment id
	DeploymentEnvironmentID string `json:"env_id"`
	// Plan holds the composited rule set about the pilot participation decision logic.
	RolloutPlan RolloutDefinitionView `json:"plan"`
}

func (r Rollout) MarshalJSON() ([]byte, error) {
	return json.Marshal(rolloutView{
		ID:                      r.ID,
		FlagID:                  r.FlagID,
		DeploymentEnvironmentID: r.DeploymentEnvironmentID,
		RolloutPlan:             RolloutDefinitionView{Definition: r.Plan},
	})
}

func (r *Rollout) UnmarshalJSON(bs []byte) error {
	var v rolloutView

	if err := json.Unmarshal(bs, &v); err != nil {
		return err
	}

	r.ID = v.ID
	r.FlagID = v.FlagID
	r.DeploymentEnvironmentID = v.DeploymentEnvironmentID
	r.Plan = v.RolloutPlan.Definition
	return nil
}
