package main

type DBStorage interface {
	Init() (DBStorage, error)
	GetValue(key string) (string, error)
	SetValue(key string, value interface{}) error
}
