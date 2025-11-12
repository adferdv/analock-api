package services

import (
	"testing"
	"time"

	"github.com/adfer-dev/analock-api/models"
)

var cacheService = &cacheServiceImpl{cache: newCache(5*time.Minute, 1*time.Minute)}

func TestCacheResource(t *testing.T) {
	cacheService.CacheResource(func() (interface{}, error) {
		return models.DiaryEntry{
			Id:      1,
			Title:   "",
			Content: "",
			Registration: models.ActivityRegistration{
				Id:               1,
				RegistrationDate: 123,
				UserRefer:        1,
			},
		}, nil
	}, "diaryEntries", "user-1")

	_, err := cacheService.cache.get("diaryEntries-user-1")

	if err != nil {
		t.Fatal("Entry is not cached")
	}
}

func TestEvictCacheResource(t *testing.T) {
	cacheService.EvictResourceItem("diaryEntries", "user-1")

	_, err := cacheService.cache.get("diaryEntries-user-1")

	if err == nil {
		t.Fatal("Entry is still cached")
	}
}
