package redis

import (
	"github.com/evassilyev/secret-server/api/core"
)

func NewSecretService(addr, pass string, db int) (core.SecretService, error) {
	return &secretService{}, nil
}

type secretService struct {
}

func (*secretService) Save(secret string, eav, ea int) (core.Secret, error) {
	panic("implement me")
}

func (*secretService) Get(hash string) (core.Secret, error) {
	panic("implement me")
}
