package internal_models

import "github.com/gbrlsnchs/jwt/v3"

type LoginToken struct {
	jwt.Payload
	ID       string `json:"id"`
	Username string `json:"username"`
}
