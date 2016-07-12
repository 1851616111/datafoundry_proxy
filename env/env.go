package env

type Env interface {
	Init()
	Validate(func(k string))
	Get(name string) string
	Set(key, value string)
	Print()
}