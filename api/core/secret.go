package core

import (
	"crypto/md5"
	"fmt"
	"time"
)

type Secret struct {
	Hash           string
	SecretText     string
	CreatedAt      string
	ExpiresAt      string
	RemainingViews int
}

type SecretService interface {
	Save(secret string, eav, ea int) (Secret, error)
	Get(hash string) (Secret, error)
}

func NewSecret(secret string, eav, ea int) Secret {
	var res Secret
	now := time.Now()
	res.Hash = fmt.Sprintf("%x", md5.Sum([]byte(secret)))
	res.SecretText = secret
	res.RemainingViews = eav
	res.CreatedAt = now.String()
	res.ExpiresAt = now.Add(time.Duration(ea) * time.Minute).String()
	return res
}
