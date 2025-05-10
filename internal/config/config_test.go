package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type StoreContentCachedStaleTimeTestSuite struct {
	suite.Suite
}

func (s *StoreContentCachedStaleTimeTestSuite) TestStoreContentCachedStaleTime() {
	_, err := parseStoreContentCachedStaleTime("*:12h:4h")
	s.ErrorContains(err, "must be at least 18h")

	_, err = parseStoreContentCachedStaleTime("*:18h:4h")
	s.ErrorContains(err, "must be at least 6h")

	_, err = parseStoreContentCachedStaleTime("*:1d:8h")
	s.ErrorContains(err, "invalid")

	staleTime, err := parseStoreContentCachedStaleTime("*:36h:12h")
	s.Nil(err)
	s.Equal(staleTime.GetStaleTime(true, "realdebrid"), 36*time.Hour)
	s.Equal(staleTime.GetStaleTime(false, "realdebrid"), 12*time.Hour)
	s.Equal(staleTime.GetStaleTime(true, "torbox"), 36*time.Hour)
	s.Equal(staleTime.GetStaleTime(false, "torbox"), 12*time.Hour)

	staleTime, err = parseStoreContentCachedStaleTime("*:36h:12h,realdebrid:48h:16h")
	s.Nil(err)
	s.Equal(staleTime.GetStaleTime(true, "realdebrid"), 48*time.Hour)
	s.Equal(staleTime.GetStaleTime(false, "realdebrid"), 16*time.Hour)
	s.Equal(staleTime.GetStaleTime(true, "torbox"), 36*time.Hour)
	s.Equal(staleTime.GetStaleTime(false, "torbox"), 12*time.Hour)

	staleTime, err = parseStoreContentCachedStaleTime("realdebrid:48h:16h")
	s.Nil(err)
	s.Equal(staleTime.GetStaleTime(true, "realdebrid"), 48*time.Hour)
	s.Equal(staleTime.GetStaleTime(false, "realdebrid"), 16*time.Hour)
	s.Equal(staleTime.GetStaleTime(true, "torbox"), 24*time.Hour)
	s.Equal(staleTime.GetStaleTime(false, "torbox"), 8*time.Hour)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(StoreContentCachedStaleTimeTestSuite))
}
