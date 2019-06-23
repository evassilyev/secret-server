package pgdb

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/evassilyev/secret-server/api/core"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
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

type dbSecret struct {
	Id sql.NullString `db:"id"`
	core.Secret
}

func (dbs *dbSecret) toCoreSecret() core.Secret {
	return core.Secret{
		Hash:           dbs.Hash,
		SecretText:     dbs.SecretText,
		CreatedAt:      dbs.CreatedAt,
		ExpiresAt:      dbs.ExpiresAt,
		RemainingViews: dbs.RemainingViews,
	}
}

func (s *secretService) Save(secret string, eav, ea int) (res core.Secret, err error) {

	newSecret := dbSecret{
		Id: sql.NullString{String: uuid.New().String(), Valid: true},
	}
	newSecret.Secret = core.NewSecret(secret, eav, ea)

	insq := `INSERT INTO secrets(id, hash, secret_text, created_at, expires_at, remaining_views) 
		VALUES (:id, :hash, :secret_text, :created_at, :expires_at, :remaining_views)`
	_, err = s.db.NamedExec(insq, newSecret)
	if err != nil {
		return
	}

	selq := `SELECT hash, secret_text, created_at, expires_at, remaining_views FROM secrets WHERE id=$1 LIMIT 1`
	err = s.db.Get(&res, selq, newSecret.Id)
	return
}

func (s *secretService) Get(hash string) (secret core.Secret, err error) {

	var tx *sqlx.Tx
	tx, err = s.db.Beginx()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var dbs dbSecret
	q := `UPDATE secrets SET remaining_views = remaining_views - 1  WHERE hash=$1 
		RETURNING id, hash, secret_text, created_at, expires_at, remaining_views`
	err = tx.Get(&dbs, q, hash)
	if err != nil {
		fmt.Println(err)
		return
	}

	exptime, err := time.Parse(time.RFC3339, dbs.ExpiresAt)
	if err != nil {
		fmt.Println(err)
		return
	}

	expired := exptime.Before(time.Now())

	if dbs.RemainingViews <= 0 || expired {
		deleteq := `DELETE FROM secrets WHERE id=$1`
		_, err = tx.Exec(deleteq, dbs.Id)
		if err != nil {
			return
		}
		if expired {
			err = errors.New("secret expired")
			return
		}
	}

	secret = dbs.toCoreSecret()
	return
}
