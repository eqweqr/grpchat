package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type MessageStore interface {
	SetMessage(ctx context.Context, message string, expireIn time.Duration) (string, error)
	DeleteMessage(ctx context.Context, messageId string) error
	GetMessage(ctx context.Context, messageId string) (any, error)
}

type RedisMessageStore struct {
	RedisClient *redis.Client
}

func NewRedisMessageStore(address string) (*RedisMessageStore, error) {
	log.Printf("Connecting to redis: %s", address)
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("error connecting to redis: %w", err)
	}
	return &RedisMessageStore{RedisClient: rdb}, nil
}

func (store *RedisMessageStore) SetMessage(ctx context.Context, message string, expireIn time.Duration) (string, error) {
	messageID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("cannot generate id: %w", err)
		return "", fmt.Errorf("cannot generate id for message: %w", err)
	}
	// uuid.Must(uuid.FromBytes([]byte("69359037-9599-48e7-b8f2-48393c019135")))
	messageIDString := messageID.String()

	if err = store.RedisClient.Set(ctx, messageIDString, message, expireIn).Err(); err != nil {
		log.Printf("cannot set message: %s", message)
		return "", fmt.Errorf("cannot insert message: %s", message)
	}
	return messageIDString, nil
}

func (store *RedisMessageStore) DeleteMessage(ctx context.Context, messageId string) error {
	if err := store.RedisClient.Del(ctx, messageId).Err(); err != nil {
		log.Printf("cannot delete messaeg(%s) from rdb", messageId)
		return fmt.Errorf("cannot delete message")
	}
	return nil
}

func (store *RedisMessageStore) GetMessage(ctx context.Context, messageId string) (any, error) {
	value, err := store.RedisClient.Get(ctx, messageId).Result()
	if err != nil {
		return nil, fmt.Errorf("cannot get message from rdb: %w", err)
	}
	return value, nil
}
