package security

import "time"

type Token struct {
	ID       string `ext:"ID"`
	SHA512   string
	OwnerUID string
	IssuedAt time.Time
	Duration time.Duration
}

func (token Token) IsValid() bool {
	now := time.Now()

	if now.Before(token.IssuedAt) {
		return false
	}

	if !token.IsExpirable() {
		return true
	}

	expiresAt := token.IssuedAt.Add(token.Duration)
	return now.Before(expiresAt)
}

// IsExpirable states whether or not the token may expire; capable of being brought to an end.
//
// p.s.: expirable is a valid word since 1913.
func (token Token) IsExpirable() bool {
	return token.Duration > 0
}
