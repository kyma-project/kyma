package db

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

type DB struct {
	client *redis.Client
}

func New(redisURL string) *DB {
	client := redis.NewClient(
		&redis.Options{
			Addr:     redisURL,
			Password: "",
			DB:       0,
		})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	if err != nil {
		log.Panicf("failed to connect to Redis %s", redisURL)
	}
	return &DB{client: client}
}
