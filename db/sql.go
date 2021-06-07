package db

import (
	"context"
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
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
	var id int64
	var err error
	shortURL := &record.ShortURL{
		CreatedAt: time.Now().Round(time.Second),
		ExpireAt:  expireAt,
		URL:       url,
		IsDeleted: false,
	}
	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRow("SELECT id FROM url_shortener.recyclable_urls LIMIT 1")
	err = row.Scan(&id)
	if err == nil {
		// recycle urls from recyclable_urls table
		shortURL.ID = id
		if _, err := tx.Exec("DELETE FROM url_shortener.recyclable_urls WHERE id = ?", id); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(
			"UPDATE url_shortener.short_urls SET url = ?, created_at = ?, expire_at = ?, is_deleted = false WHERE id = ?",
			url, shortURL.CreatedAt, expireAt, id); err != nil {
			return nil, err
		}
	} else {
		var result sql.Result
		result, err = tx.Exec("INSERT INTO url_shortener.short_urls (url, expire_at) VALUES (?, ?)", url, expireAt)
		if err != nil {
			return nil, err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return nil, err
		}
		shortURL.ID = id
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return shortURL, nil
}

func (s *sqlStore) Get(ctx context.Context, id int64) (*record.ShortURL, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, url, created_at, expire_at, is_deleted FROM url_shortener.short_urls WHERE id = ?", id)

	shortURL := &record.ShortURL{}
	if err := row.Scan(
		&shortURL.ID,
		&shortURL.URL,
		&shortURL.CreatedAt,
		&shortURL.ExpireAt,
		&shortURL.IsDeleted,
	); err != nil {
		return nil, err
	}
	return shortURL, nil
}

func (s *sqlStore) GetExpiredIDs(ctx context.Context) (<-chan int64, error) {
	now := time.Now().Round(time.Second)
	rows, err := s.db.QueryContext(ctx, "SELECT id FROM url_shortener.short_urls WHERE expire_at < ?", now)
	if err != nil {
		return nil, err
	}
	ch := make(chan int64)
	go func() {
		defer close(ch)
		defer rows.Close()
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return
			}
			ch <- id
		}
	}()
	return ch, nil
}

func (s *sqlStore) Expire(ctx context.Context, id int64) error {
	return s.delete(ctx, id, true)
}

func (s *sqlStore) Delete(ctx context.Context, id int64) error {
	return s.delete(ctx, id, false)
}

func (s *sqlStore) delete(ctx context.Context, id int64, onExpire bool) error {
	var tx *sql.Tx
	var err error

	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Errorf("")
		return err
	}
	defer tx.Rollback()

	row := tx.QueryRow("SELECT id FROM url_shortener.recyclable_urls WHERE id = ?", id)
	var recyclableID int64
	err = row.Scan(&recyclableID)
	if err == nil {
		return nil
	}
	if err != ErrNoRows {
		return err
	}

	if onExpire {
		row = tx.QueryRow("SELECT id FROM url_shortener.short_urls WHERE id = ? AND expire_at < ?",
			id, time.Now().Round(time.Second))
	} else {
		row = tx.QueryRow("SELECT id FROM url_shortener.short_urls WHERE id = ?", id)
	}

	var shortID int64
	err = row.Scan(&shortID)
	if err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO url_shortener.recyclable_urls (id) VALUES (?)", id); err != nil {
		return err
	}
	return tx.Commit()
}
