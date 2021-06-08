package converter

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestConverter(t *testing.T) {
	suite.Run(t, new(ConverterTestSuite))
}

type ConverterTestSuite struct {
	suite.Suite

	conv *converterImpl
}

func (s *ConverterTestSuite) SetupTest() {
	s.conv = NewConverter()
}

func (s *ConverterTestSuite) TestConvertToID() {
	urlID := "67890"

	gotID, gotErr := s.conv.ConvertToID(urlID)

	s.NoError(gotErr)
	s.Equal(int64(67890), gotID)
}

func (s *ConverterTestSuite) TestConvertToID_withFormatError() {
	urlID := "abcde"

	gotID, gotErr := s.conv.ConvertToID(urlID)

	s.Equal(int64(0), gotID)
	s.Error(gotErr)
	s.Equal(gotErr, ErrURLFormat)
}

func (s *ConverterTestSuite) TestConvertToShortURL() {
	id := int64(12345)

	gotShortURL, gotErr := s.conv.ConvertToURLID(id)

	s.NoError(gotErr)
	s.Equal("12345", gotShortURL)
}
