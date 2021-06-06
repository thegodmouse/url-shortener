package db

import (
	"context"
	"database/sql"
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
	expRows := sqlmock.NewRows([]string{"id", "url", "created_at", "expire_at", "is_deleted"}).
		AddRow(id, url, createdAt, expireAt, false)
	s.mock.
		ExpectBegin()
	s.mock.
		ExpectExec("INSERT INTO url_shortener.short_urls \\(url, expire_at\\) VALUES \\(\\?, \\?\\)").
		WithArgs(url, expireAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.
		ExpectQuery(
			"SELECT id, url, created_at, expire_at, is_deleted FROM url_shortener\\.short_urls").
		WithArgs(id).
		WillReturnRows(expRows)
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

func (s *SQLTestSuite) TestDelete() {
	sqlStore := NewSQLStore(s.db)

	id := int64(12345)

	s.mock.
		ExpectExec("UPDATE url_shortener.short_urls SET is_deleted = true WHERE id = \\?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// SUT
	gotErr := sqlStore.Delete(context.Background(), id)

	s.NoError(gotErr)
}
