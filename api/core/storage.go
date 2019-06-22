package core

type Secret struct {
	Hash           string
	SecretText     string
	CreatedAt      string
	ExpiresAt      string
	RemainingViews int32
}

type StorageService interface {
	Save(secret Secret) error
	Get(hash string) (Secret, error)
}
