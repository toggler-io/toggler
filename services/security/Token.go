package security

import "time"

type Token struct {
	ID       string `ext:"ID"`
	Token    string
	UserUID  string
	IssuedAt time.Time
	Duration time.Duration
}

func (token Token) IsValid() bool {
	now := time.Now()

	if now.Before(token.IssuedAt) {
		return false
	}

	if token.Duration == 0 {
		return true
	}

	expiresAt := token.IssuedAt.Add(token.Duration)
	return now.Before(expiresAt)
}
