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
	str, err := db.client.Get(key).Result()
	if err != nil {
		log.Printf("got error {%+v} when retrieving for the key {%s}", err, key)
		return "", err
	}

	log.Printf("Got redis value {%s} for key {%s}", str, key)
	return str, nil
}
