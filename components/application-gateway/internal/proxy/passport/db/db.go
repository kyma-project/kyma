package db

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
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

func (db *DB) GET(key string) (string, error) {
	strCmd := db.client.Get(key)
	if err := strCmd.Err(); err != nil {
		log.Printf("got error when retrieving for the key {%s}", key)
		return "", err
	}
	return strCmd.Val(), nil
}
