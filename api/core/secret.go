package core

import (
	"crypto/md5"
	"fmt"
	"time"
)

type Secret struct {
	Hash           string `json:"hash" db:"hash"`
	SecretText     string `json:"secretText" db:"secret_text"`
	CreatedAt      string `json:"createdAt" db:"created_at"`
	ExpiresAt      string `json:"expiresAt" db:"expires_at"`
	RemainingViews int    `json:"remainingViews" db:"remaining_views"`
}

type SecretService interface {
	Save(secret string, eav, ea int) (Secret, error)
	Get(hash string) (Secret, error)
}

func NewSecret(secret string, eav, ea int) Secret {
	// TODO case ea == 0
	var res Secret
	now := time.Now()
	res.Hash = fmt.Sprintf("%x", md5.Sum([]byte(secret)))
	res.SecretText = secret
	res.RemainingViews = eav
	res.CreatedAt = now.String()
	res.ExpiresAt = now.Add(time.Duration(ea) * time.Minute).String()
	return res
}
