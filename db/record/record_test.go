package record

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecord(t *testing.T) {
	r1 := &ShortURL{
		ID:        123,
		CreatedAt: time.Now().Add(-time.Hour).Round(time.Second),
		ExpireAt:  time.Now().Add(time.Hour).Round(time.Second),
		URL:       "http://localhost:23456",
		IsDeleted: false,
	}

	// SUT
	gotBR1, gotErr := r1.MarshalBinary()
	assert.NoError(t, gotErr)

	r2 := &ShortURL{}
	gotErr = r2.UnmarshalBinary(gotBR1)
	assert.NoError(t, gotErr)

	assert.Equal(t, r1.ID, r2.ID)
	assert.Equal(t, r1.URL, r2.URL)
	assert.Equal(t, r1.CreatedAt, r2.CreatedAt)
	assert.Equal(t, r1.ExpireAt, r2.ExpireAt)
	assert.Equal(t, r1.IsDeleted, r2.IsDeleted)
}
