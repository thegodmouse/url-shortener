package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

func TestSQLSuite(t *testing.T) {
	suite.Run(t, new(SQLTestSuite))
}

type SQLTestSuite struct {
	suite.Suite

	db   *sql.DB
	mock sqlmock.Sqlmock
}

func (s *SQLTestSuite) SetupTest() {
	var err error
	s.db, s.mock, err = sqlmock.New()
	if err != nil {
		panic(err)
	}
}

func (s *SQLTestSuite) TearDownTest() {
	s.mock.ExpectationsWereMet()
	s.db.Close()
}

func (s *SQLTestSuite) TestCreate() {
	sqlStore := NewSQLStore(s.db)

	id := int64(1)
	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectExec("INSERT INTO url_shortener.short_urls \\(url, expire_at\\) VALUES \\(\\?, \\?\\)").
		WithArgs(url, expireAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectCommit()

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.NoError(gotErr)
	s.Equal(id, gotRecord.ID)
	s.Equal(url, gotRecord.URL)
	s.Equal(expireAt, gotRecord.ExpireAt)
	s.False(gotRecord.IsDeleted)
}

func (s *SQLTestSuite) TestCreate_withRecyclableURL() {
	sqlStore := NewSQLStore(s.db)

	id := int64(1)
	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)
	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec(
			"DELETE FROM url_shortener.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectExec(
			"UPDATE url_shortener.short_urls SET url = \\?, created_at = \\?, expire_at = \\?, is_deleted = false WHERE id = \\?").
		WithArgs(url, createdAt, expireAt, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectCommit()

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.NoError(gotErr)
	s.Equal(id, gotRecord.ID)
	s.Equal(url, gotRecord.URL)
	s.Equal(expireAt, gotRecord.ExpireAt)
	s.False(gotRecord.IsDeleted)
}

func (s *SQLTestSuite) TestCreate_withBeginError() {
	sqlStore := NewSQLStore(s.db)

	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)

	s.mock.
		ExpectBegin().
		WillReturnError(errors.New("unknown begin error"))

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.Error(gotErr)
	s.Nil(gotRecord)
}

func (s *SQLTestSuite) TestCreate_withQueryRecyclableError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(1)
	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnError(errors.New("unknown query error"))
	s.mock.
		ExpectExec("INSERT INTO url_shortener.short_urls \\(url, expire_at\\) VALUES \\(\\?, \\?\\)").
		WithArgs(url, expireAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectCommit()

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.NoError(gotErr)
	s.Equal(id, gotRecord.ID)
	s.Equal(url, gotRecord.URL)
	s.Equal(expireAt, gotRecord.ExpireAt)
	s.False(gotRecord.IsDeleted)
}

func (s *SQLTestSuite) TestCreate_withDeleteRecyclableError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(1)
	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec(
			"DELETE FROM url_shortener.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(errors.New("unknown delete error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.Error(gotErr)
	s.Nil(gotRecord)
}

func (s *SQLTestSuite) TestCreate_withUpdateShortError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(1)
	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec(
			"DELETE FROM url_shortener.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectExec(
			"UPDATE url_shortener.short_urls SET url = \\?, created_at = \\?, expire_at = \\?, is_deleted = false WHERE id = \\?").
		WithArgs(url, createdAt, expireAt, id).
		WillReturnError(errors.New("unknown update error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.Error(gotErr)
	s.Nil(gotRecord)
}

func (s *SQLTestSuite) TestCreate_withInsertShortError() {
	sqlStore := NewSQLStore(s.db)

	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectExec("INSERT INTO url_shortener.short_urls \\(url, expire_at\\) VALUES \\(\\?, \\?\\)").
		WithArgs(url, expireAt).
		WillReturnError(errors.New("unknown insert error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.Error(gotErr)
	s.Nil(gotRecord)
}

func (s *SQLTestSuite) TestCreate_withGetLastInsertIDError() {
	sqlStore := NewSQLStore(s.db)

	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)
	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectExec("INSERT INTO url_shortener.short_urls \\(url, expire_at\\) VALUES \\(\\?, \\?\\)").
		WithArgs(url, expireAt).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("unknown result error")))
	s.mock.
		ExpectRollback()
	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.Error(gotErr)
	s.Nil(gotRecord)
}

func (s *SQLTestSuite) TestCreate_withCommitError() {
	sqlStore := NewSQLStore(s.db)

	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls LIMIT 1 FOR UPDATE SKIP LOCKED").
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectExec("INSERT INTO url_shortener.short_urls \\(url, expire_at\\) VALUES \\(\\?, \\?\\)").
		WithArgs(url, expireAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectCommit().
		WillReturnError(errors.New("unknown commit error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotRecord, gotErr := sqlStore.Create(context.Background(), url, expireAt)

	s.Error(gotErr)
	s.Nil(gotRecord)
}

func (s *SQLTestSuite) TestGet() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)
	url := "http://localhost:5566"
	createdAt := time.Now().Round(time.Second)
	expireAt := createdAt.Add(time.Minute).Round(time.Second)
	expRows := sqlmock.NewRows([]string{"id", "url", "created_at", "expire_at", "is_deleted"}).
		AddRow(id, url, createdAt, expireAt, false)

	s.mock.
		ExpectQuery("SELECT id, url, created_at, expire_at, is_deleted " +
			"FROM url_shortener\\.short_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnRows(expRows)

	// SUT
	gotRecord, gotErr := sqlStore.Get(context.Background(), id)

	s.NoError(gotErr)
	s.Equal(id, gotRecord.ID)
	s.Equal(url, gotRecord.URL)
	s.Equal(expireAt, gotRecord.ExpireAt)
	s.False(gotRecord.IsDeleted)
}

func (s *SQLTestSuite) TestGet_withQueryShortError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectQuery("SELECT id, url, created_at, expire_at, is_deleted " +
			"FROM url_shortener\\.short_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(errors.New("unknown query error"))

	// SUT
	gotRecord, gotErr := sqlStore.Get(context.Background(), id)

	s.Error(gotErr)
	s.Nil(gotRecord)
}

func (s *SQLTestSuite) TestDelete() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.short_urls WHERE id = \\? FOR UPDATE").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = \\?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectExec("INSERT INTO url_shortener.recyclable_urls \\(id\\) VALUES \\(\\?\\)").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectCommit()
	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.NoError(gotErr)
}

func (s *SQLTestSuite) TestDelete_withAlreadyDeleted() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)
	expRows := sqlmock.NewRows([]string{"id"}).AddRow(id)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnRows(expRows)
	s.mock.
		ExpectCommit()

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.NoError(gotErr)
}

func (s *SQLTestSuite) TestDelete_withBeginTransactionError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin().
		WillReturnError(errors.New("unknown begin error"))

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *SQLTestSuite) TestDelete_withQueryRecyclableError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(errors.New("unknown query error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *SQLTestSuite) TestDelete_withQueryShortError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.short_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(errors.New("unknown query error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *SQLTestSuite) TestDelete_withUpdateShortError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.short_urls WHERE id = \\? FOR UPDATE").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = \\?").
		WithArgs(id).
		WillReturnError(errors.New("unknown update error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *SQLTestSuite) TestDelete_withInsertRecyclableError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.short_urls WHERE id = \\? FOR UPDATE").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = \\?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectExec("INSERT INTO url_shortener.recyclable_urls \\(id\\) VALUES \\(\\?\\)").
		WithArgs(id).
		WillReturnError(errors.New("unknown insert error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *SQLTestSuite) TestDelete_withCommitError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.short_urls WHERE id = \\? FOR UPDATE").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = \\?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectExec("INSERT INTO url_shortener.recyclable_urls \\(id\\) VALUES \\(\\?\\)").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectCommit().
		WillReturnError(errors.New("unknown commit error"))
	s.mock.
		ExpectRollback()

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *SQLTestSuite) TestExpire() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.short_urls WHERE id = \\? AND expire_at < \\? FOR UPDATE").
		WithArgs(id, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	s.mock.
		ExpectExec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = \\?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectExec("INSERT INTO url_shortener.recyclable_urls \\(id\\) VALUES \\(\\?\\)").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectCommit()
	// SUT
	gotErr := sqlStore.Expire(context.Background(), id)

	s.NoError(gotErr)
}

func (s *SQLTestSuite) TestExpire_withError() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectBegin()
	s.mock.
		ExpectQuery("SELECT id FROM url_shortener\\.recyclable_urls WHERE id = \\?").
		WithArgs(id).
		WillReturnError(errors.New("unknown query error"))
	s.mock.
		ExpectRollback()
	// SUT
	gotErr := sqlStore.Expire(context.Background(), id)

	s.Error(gotErr)
}

func (s *SQLTestSuite) TestGetExpiredIDs() {
	sqlStore := NewSQLStore(s.db)

	expRows := sqlmock.NewRows([]string{"id"}).
		AddRow(int64(1)).
		AddRow(int64(2)).
		AddRow(int64(3))

	s.mock.
		ExpectQuery("SELECT id FROM url_shortener.short_urls WHERE expire_at < \\? AND is_deleted = false").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(expRows)

	gotCh, gotErr := sqlStore.GetExpiredIDs(context.Background())

	s.NoError(gotErr)
	var result []int64
	for id := range gotCh {
		result = append(result, id)
	}
	s.Contains(result, int64(1))
	s.Contains(result, int64(2))
	s.Contains(result, int64(3))
	s.Len(result, 3)
}

func (s *SQLTestSuite) TestGetExpiredIDs_withQueryError() {
	sqlStore := NewSQLStore(s.db)

	s.mock.
		ExpectQuery("SELECT id FROM url_shortener.short_urls WHERE expire_at < \\? AND is_deleted = false").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("unknown query error"))

	gotCh, gotErr := sqlStore.GetExpiredIDs(context.Background())

	s.Error(gotErr)
	s.Nil(gotCh)
}

func (s *SQLTestSuite) TestGetExpiredIDs_withRowsError() {
	sqlStore := NewSQLStore(s.db)

	expRows := sqlmock.NewRows([]string{"id"}).
		AddRow(int64(1)).
		AddRow(int64(2)).
		AddRow(int64(3)).
		RowError(2, errors.New("scan error after second row"))

	s.mock.
		ExpectQuery("SELECT id FROM url_shortener.short_urls WHERE expire_at < \\? AND is_deleted = false").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(expRows)

	gotCh, gotErr := sqlStore.GetExpiredIDs(context.Background())

	s.NoError(gotErr)
	var result []int64
	for id := range gotCh {
		result = append(result, id)
	}
	s.Contains(result, int64(1))
	s.Contains(result, int64(2))
	s.NotContains(result, int64(3))
	s.Len(result, 2)
}

func (s *SQLTestSuite) TestGetExpiredIDs_withScanError() {
	sqlStore := NewSQLStore(s.db)

	expRows := sqlmock.NewRows([]string{"id"}).
		AddRow(int64(1)).
		AddRow("invalid-id")

	s.mock.
		ExpectQuery("SELECT id FROM url_shortener.short_urls WHERE expire_at < \\? AND is_deleted = false").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(expRows)

	gotCh, gotErr := sqlStore.GetExpiredIDs(context.Background())

	s.NoError(gotErr)
	var result []int64
	for id := range gotCh {
		result = append(result, id)
	}
	s.Len(result, 1)
}
