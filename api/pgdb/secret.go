package pgdb

import (
	"github.com/evassilyev/secret-server/api/core"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewSecretService(url string, mic, moc int) (core.SecretService, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(mic)
	db.SetMaxOpenConns(moc)

	return &secretService{db: db}, nil
}

type secretService struct {
	db *sqlx.DB
}

func (s *secretService) Save(secret string, eav, ea int) (core.Secret, error) {

	type dbSecret struct {
		Id string `db:"id"`
		core.Secret
	}

	//TODO refactor it
	newSecret := core.NewSecret(secret, eav, ea)
	newSecretDb := dbSecret{
		Id: uuid.New().String(),
	}
	newSecretDb.Hash = newSecret.Hash
	newSecretDb.SecretText = newSecret.SecretText
	newSecretDb.CreatedAt = newSecret.CreatedAt
	newSecretDb.ExpiresAt = newSecret.ExpiresAt
	newSecretDb.RemainingViews = newSecret.RemainingViews

	q := `INSERT INTO secrets(id, hash, secret_text, created_at, expires_at, remaining_views) 
		VALUES (:id, :hash, :secret_text, :created_at, :expires_at, :remaining_views)`
	_, err := s.db.NamedExec(q, newSecretDb)
	if err != nil {
		return core.Secret{}, err
	}

	return s.getById(newSecretDb.Id)
}

func (s *secretService) Get(hash string) (secret core.Secret, err error) {
	q := `UPDATE secrets SET remaining_views = remaining_views - 1  WHERE hash=$1 
		RETURNING hash, secret_text, created_at, expires_at, remaining_views`
	return s.get(q, hash)
}

func (s *secretService) getById(id string) (core.Secret, error) {
	q := `SELECT hash, secret_text, created_at, expires_at, remaining_views FROM secrets WHERE id=$1 LIMIT 1`
	return s.get(q, id)
}

func (s *secretService) get(query, searchParameter string) (core.Secret, error) {
	var secret core.Secret
	err := s.db.Get(&secret, query, searchParameter)
	return secret, err
}
