package core

import (
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
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
	var res Secret
	now := time.Now()
	res.Hash = fmt.Sprintf("%x", md5.Sum([]byte(uuid.New().String())))
	res.SecretText = secret
	res.RemainingViews = eav
	res.CreatedAt = now.Format(time.RFC3339)
	if ea == 0 {
		res.ExpiresAt = "9999-12-31T23:59:59Z" //time.RFC3339 format
	} else {
		res.ExpiresAt = now.Add(time.Duration(ea) * time.Minute).Format(time.RFC3339)
	}
	return res
}
