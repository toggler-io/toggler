package security

import "time"

type Issuer struct {
	Storage
}

func (i *Issuer) CreateNewToken(userUID string, issueAt *time.Time, duration *time.Duration) (*Token, error) {
	return nil, nil
}
