package service

import "context"

type MessageStore interface {
	Save(ctx context.Context, message string, auth string) error
}
