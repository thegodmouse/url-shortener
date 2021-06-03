package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/thegodmouse/url-shortener/db/record"
)

func NewSQLStore(db *sql.DB) *sqlStore {
	return &sqlStore{
		db: db,
	}
}

type sqlStore struct {
	db *sql.DB
}

func (s *sqlStore) Create(ctx context.Context, url string, expireAt time.Time) (*record.ShortURL, error) {

	var tx *sql.Tx
	var result sql.Result
	var id int64
	var err error

	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err = tx.Exec(
		"INSERT INTO url_shortener.short_urls (url, expire_at) VALUES (?, ?)",
		url, expireAt)
	if err != nil {
		return nil, err
	}
	id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow("SELECT * FROM url_shortener.short_urls WHERE id = ?", id)

	var shortURL *record.ShortURL
	err = row.Scan(shortURL)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return shortURL, nil
}

func (s *sqlStore) Get(ctx context.Context, id int64) (*record.ShortURL, error) {
	panic("implement me")
}

func (s *sqlStore) Delete(ctx context.Context, id int64) error {
	panic("implement me")
}
