package db

import (
	"context"
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/db/record"
)

// NewSQLStore returns a new db.Store which is implemented by sql database.
func NewSQLStore(db *sql.DB) *sqlStore {
	return &sqlStore{
		db: db,
	}
}

type sqlStore struct {
	db *sql.DB
}

// Create creates a new short url record or recycles an old one from expired or deleted records.
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
		log.Errorf("sqlStore.Create: begin transaction err: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRow("SELECT id FROM url_shortener.recyclable_urls LIMIT 1 FOR UPDATE")
	err = row.Scan(&id)
	if err == nil {
		// recycle urls from recyclable_urls table
		shortURL.ID = id
		if _, err := tx.Exec("DELETE FROM url_shortener.recyclable_urls WHERE id = ?", id); err != nil {
			log.Errorf("sqlStore.Create: delete recyclable url err: %v, with id: %v", err, id)
			return nil, err
		}
		if _, err := tx.Exec(
			"UPDATE url_shortener.short_urls SET url = ?, created_at = ?, expire_at = ?, is_deleted = false WHERE id = ?",
			url, shortURL.CreatedAt, expireAt, id); err != nil {
			log.Errorf("sqlStore.Create: query recyclable url err: %v, with id: %v", err, id)
			return nil, err
		}
		log.Infof("sqlStore.Create: use the recycle url record with id: %v", id)
	} else {

		var result sql.Result
		result, err = tx.Exec("INSERT INTO url_shortener.short_urls (url, expire_at) VALUES (?, ?)", url, expireAt)
		if err != nil {
			log.Errorf("sqlStore.Create: insert new sql record err: %v, with url: %v", err, url)
			return nil, err
		}
		id, err = result.LastInsertId()
		if err != nil {
			log.Errorf("sqlStore.Create: get results from query err: %v, with url: %v", err, url)
			return nil, err
		}
		shortURL.ID = id
		log.Infof("sqlStore.Create: use the new created url record with id: %v", id)
	}
	if err := tx.Commit(); err != nil {
		log.Errorf("sqlStore.Create: unable to commit changes for the transaction")
		return nil, err
	}
	log.Infof("sqlStore.Create: successfully create or recycle an url record with id: %v", id)
	return shortURL, nil
}

// Get gets the short url record with the given id.
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
		log.Errorf("sqlStore.Get: query url record err: %v, with id: %v", err, id)
		return nil, err
	}
	log.Infof("sqlStore.Get: successfully get url record with id: %v", id)
	return shortURL, nil
}

// GetExpiredIDs returns a channel for reading expired ids.
func (s *sqlStore) GetExpiredIDs(ctx context.Context) (<-chan int64, error) {
	now := time.Now().Round(time.Second)
	rows, err := s.db.QueryContext(ctx, "SELECT id FROM url_shortener.short_urls WHERE expire_at < ? AND is_deleted = false", now)
	if err != nil {
		log.Errorf("sqlStore.GetExpiredIDs: query expired ids err: %v", err)
		return nil, err
	}
	ch := make(chan int64)
	go func() {
		defer close(ch)
		defer rows.Close()
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				log.Errorf("sqlStore.GetExpiredIDs: scan for row err: %v", err)
				return
			}
			log.Infof("sqlStore.GetExpiredIDs: record is expired with id: %v", id)
			ch <- id
		}
	}()
	return ch, nil
}

// Expire expires the short url record with the given id, and makes it recyclable.
func (s *sqlStore) Expire(ctx context.Context, id int64) error {
	if err := s.delete(ctx, id, true); err != nil {
		log.Errorf("sqlStore.Expire: expire record err: %v, with id: %v", err, id)
		return err
	}
	log.Infof("sqlStore.Expire: finished with id: %v", id)
	return nil
}

// Delete deletes the short url record with the given id, and makes is recyclable.
func (s *sqlStore) Delete(ctx context.Context, id int64) error {
	if err := s.delete(ctx, id, false); err != nil {
		log.Errorf("sqlStore.Delete: delete record err: %v, with id: %v", err, id)
		return err
	}
	log.Infof("sqlStore.Delete: finished with id: %v", id)
	return nil
}

func (s *sqlStore) delete(ctx context.Context, id int64, onExpire bool) error {
	var tx *sql.Tx
	var err error

	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Errorf("sqlStore.delete: begin transaction err: %v", err)
		return err
	}
	defer tx.Rollback()

	row := tx.QueryRow("SELECT id FROM url_shortener.recyclable_urls WHERE id = ? FOR UPDATE", id)
	var recyclableID int64
	err = row.Scan(&recyclableID)
	if err == nil {
		log.Infof("sqlStore.delete: url record is already deleted with id: %v", id)
		return nil
	}
	if err != ErrNoRows {
		log.Errorf("sqlStore.delete: query recyclable url err: %v, with id: %v", err, id)
		return err
	}

	if onExpire {
		row = tx.QueryRow("SELECT id FROM url_shortener.short_urls WHERE id = ? AND expire_at < ? FOR UPDATE",
			id, time.Now().Round(time.Second))
	} else {
		row = tx.QueryRow("SELECT id FROM url_shortener.short_urls WHERE id = ? FOR UPDATE", id)
	}

	var shortID int64
	err = row.Scan(&shortID)
	if err != nil {
		log.Errorf("sqlStore.delete: scan for short url id err: %v, with id: %v", err, id)
		return err
	}

	if _, err := tx.Exec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = ?", id); err != nil {
		log.Errorf("sqlStore.delete: update url as deleted err: %v with id: %v", err, id)
		return err
	}
	if _, err := tx.Exec("INSERT INTO url_shortener.recyclable_urls (id) VALUES (?)", id); err != nil {
		log.Errorf("sqlStore.delete: insert sql record to recyclable urls err: %v, with id: %v", err, id)
		return err
	}
	return tx.Commit()
}
