package internal

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	Rdb *redis.Client
	Ctx = context.Background()
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Redis not connected: %v", err)
	}
}
