package chatgpt

import (
	"github.com/go-redis/redis"
)

func NewRedis(addr, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	return client
}
