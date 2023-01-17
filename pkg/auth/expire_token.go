package auth

import "time"

type ExpireToken struct {
	Exp time.Time `json:"exp"`
}

func (e *ExpireToken) SetTTL(seconds int) {
	e.Exp = time.Now().Add(time.Second * time.Duration(seconds))
}

func (e *ExpireToken) Expired() bool {
	return time.Now().After(e.Exp)
}
