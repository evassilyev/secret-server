package dev

import (
	"fmt"

	"github.com/evassilyev/secret-server/api/core"
)

func NewSecretService() core.SecretService {
	return &secretService{}
}

type secretService struct {
}

func (secretService) Save(secret string, eav, ea int) (core.Secret, error) {
	return core.NewSecret(secret, eav, ea), nil
}

func (secretService) Get(hash string) (core.Secret, error) {
	fmt.Println(hash)
	return core.NewSecret("TESTTEST", 1, 1), nil
}
