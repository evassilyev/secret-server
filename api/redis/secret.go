package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/evassilyev/secret-server/api/core"
	"github.com/go-redis/redis"
)

func NewSecretService(addr, pass string, db int) core.SecretService {
	r := redis.NewClient(&redis.Options{
		Addr: addr, Password: pass, DB: db,
	})
	return &secretService{redis: r}
}

type secretService struct {
	redis *redis.Client
}

type secretRedisType core.Secret

func (c *secretRedisType) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *secretRedisType) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

func counterKey(hash string) string {
	return fmt.Sprintf("%s_counter", hash)
}

func (s *secretService) Save(secret string, eav, ea int) (core.Secret, error) {
	ns := core.NewSecret(secret, eav, ea)
	srt := secretRedisType(ns)

	counterKey := counterKey(ns.Hash)
	err := s.redis.Set(counterKey, eav, time.Duration(ea)*time.Minute).Err()
	if err != nil {
		return core.Secret{}, err
	}
	err = s.redis.Set(ns.Hash, &srt, time.Duration(ea)*time.Minute).Err()
	if err != nil {
		return core.Secret{}, err
	}

	cmd := s.redis.Get(ns.Hash)
	if cmd.Err() != nil {
		return core.Secret{}, cmd.Err()
	}
	var sr secretRedisType
	err = cmd.Scan(&sr)
	return core.Secret(sr), err
}

func (s *secretService) Get(hash string) (res core.Secret, err error) {

	var sr secretRedisType
	counterKey := counterKey(hash)

	err = s.redis.Watch(func(tx *redis.Tx) error {
		err := tx.Decr(counterKey).Err()
		if err != nil {
			return err
		}
		cmd := tx.Get(counterKey)
		err = cmd.Err()
		if err != nil {
			return err
		}
		remViews, err := cmd.Int()
		if err != nil {
			return err
		}

		cmd = tx.Get(hash)
		err = cmd.Err()
		if err != nil {
			return err
		}
		err = cmd.Scan(&sr)
		if err != nil {
			return err
		}
		sr.RemainingViews = remViews

		if remViews <= 0 {
			err = tx.Del(hash, counterKey).Err()
			if err != nil {
				return err
			}
		} else {
			ttl := tx.TTL(hash)
			err = ttl.Err()
			if err != nil {
				return err
			}
			err = tx.Set(hash, &sr, ttl.Val()).Err()
			if err != nil {
				return err
			}
		}

		return err
	}, counterKey, hash)

	res = core.Secret(sr)
	return
}
