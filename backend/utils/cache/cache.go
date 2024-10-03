package cache

// dragonflydb / redis client
import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

// RedisClient is a struct that holds the redis client
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new redis client

func NewRedisClient() *RedisClient {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Invalid REDIS_DB value: %v", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
	return &RedisClient{client}
}

// Set sets a key value pair in redis
func (r *RedisClient) Set(key string, value string, expiration time.Duration) error {
	err := r.client.Set(key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key: %v", err)
	}
	return nil
}

// Get gets a value from redis
func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.client.Get(key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get key: %v", err)
	}
	return val, nil
}

// Close closes the redis client
func (r *RedisClient) Close() {
	err := r.client.Close()
	if err != nil {
		log.Printf("failed to close redis client: %v", err)
	}
}
