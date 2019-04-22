package rollouts

//import (
//	"fmt"
//)

type Rollout struct {
	ID   string `ext:"ID"`
	Type string

	//Percentage int
	//URL        string
}

//func (r Rollout) Validate() error {
//	switch r.Type {
//
//	case `PERCENTAGE`:
//		return nil
//
//	case `API`:
//		return nil
//
//	default:
//		return fmt.Errorf(`unknown rollout type: %s`, r.Type)
//
//	}
//}

type FeatureFlagRolloutStrategy struct {
	ID            string `ext:"ID"`
	RolloutID     string
	FeatureFlagID string
}
