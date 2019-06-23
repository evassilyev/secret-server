package redis

import (
	"encoding/json"
	"fmt"
	"github.com/evassilyev/secret-server/api/core"
	"github.com/go-redis/redis"
	"time"
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

// TODO add transaction
func (s *secretService) Get(hash string) (res core.Secret, err error) {

	counterKey := counterKey(hash)
	err = s.redis.Decr(counterKey).Err()
	if err != nil {
		return
	}
	cmd := s.redis.Get(counterKey)
	err = cmd.Err()
	if err != nil {
		return
	}
	remViews, err := cmd.Int()
	if err != nil {
		return
	}

	var sr secretRedisType
	cmd = s.redis.Get(hash)
	err = cmd.Err()
	if err != nil {
		return
	}
	err = cmd.Scan(&sr)
	if err != nil {
		return
	}
	sr.RemainingViews = remViews

	if remViews <= 0 {
		err = s.redis.Del(hash, counterKey).Err()
		if err != nil {
			return
		}
	} else {
		ttl := s.redis.TTL(hash)
		err = ttl.Err()
		if err != nil {
			return
		}
		err = s.redis.Set(hash, &sr, ttl.Val()).Err()
		if err != nil {
			return
		}
	}
	res = core.Secret(sr)
	return
}
