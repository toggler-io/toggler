package security

func NewDoorkeeper(s Storage) *Doorkeeper {
	return &Doorkeeper{Storage: s}
}

type Doorkeeper struct {
	Storage
}

func (dk *Doorkeeper) VerifyTokenString(tokenString string) (bool, error) {
	token, err := dk.Storage.FindTokenByTokenString(tokenString)

	if token == nil {
		return false, nil
	}

	return token.IsValid(), err
}