package main

import (
	"time"
)

type DBStorage interface {
	Init() (DBStorage, error)
	GetValue(key string) (interface{}, error)
	SetValue(key string, value interface{}, dir bool) error
	SetValuebyTTL(key string, value interface{}, ttl time.Duration) error
	Delete(key string, dir bool) error
}
