package redis

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"time"
)

var ErrNoCachedCurrency = errors.New("no cached currency")

func NewCache() *redis.Client {
	cache := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return cache
}

func CloseCache(cache *redis.Client) {
	cache.Close()
}

func SaveInfo(cache *redis.Client, currency, course string) error {
	now := time.Now()
	nextDay := now.Truncate(24 * time.Hour).Add(time.Hour * 24)
	if err := cache.Set(context.Background(), currency, course, nextDay.Sub(now)).Err(); err != nil {
		return err
	}
	return nil
}

func GetInfo(cache *redis.Client, currency string) (string, error) {
	course, err := cache.Get(context.Background(), currency).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNoCachedCurrency
		}
		return "", err
	}
	return course, nil
}
