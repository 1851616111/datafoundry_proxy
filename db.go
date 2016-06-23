package main

type DBStorage interface {
	Init() (DBStorage, error)
	GetValue(key string) (interface{}, error)
	SetValue(key string, value interface{}, dir bool) error
}
