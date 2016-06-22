package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	ETCD_HTTP_ADDR string = "ETCD_HTTP_ADDR"
	ETCD_HTTP_PORT string = "ETCD_HTTP_PORT"
	ETCD_USERNAME  string = "ETCD_USERNAME"
	ETCD_PASSWORD  string = "ETCD_PASSWORD"

	LDAP_HOST_ADDR      string = "LDAP_HOST_ADDR"
	LDAP_ADMIN_USER     string = "LDAP_ADMIN_USER"
	LDAP_ADMIN_PASSWORD string = "LDAP_ADMIN_PASSWORD"
	LDAP_BASE_DN        string = "LDAP_BASE_DN" //"cn=%s,ou=Users,dc=openstack,dc=org"
)

var (
	EtcdStorageEnv Env = &EnvOnce{
		envs: map[string]string{
			ETCD_HTTP_ADDR: "127.0.0.1:2379",
			ETCD_USERNAME:  "",
			ETCD_PASSWORD:  "",
		},
	}
	LdapEnv Env = &EnvOnce{
		envs: map[string]string{
			LDAP_HOST_ADDR:      "",
			LDAP_ADMIN_USER:     "",
			LDAP_ADMIN_PASSWORD: "",
			LDAP_BASE_DN:        "",
		},
	}
	DatafoundryEnv Env = &EnvOnce{
		envs: map[string]string{"DATAFOUNDRY_HOST_ADDR": ""},
	}
	RedisEnv Env = &EnvOnce{
		envs: map[string]string{"Redis_BackingService_Name": ""},
	}
)

type Env interface {
	Init()
	Validate(func(k string))
	Get(name string) string
	Print()
}

type EnvOnce struct {
	envs map[string]string
	once sync.Once
}

func (e *EnvOnce) Init() {
	fn := func() {
		for k := range e.envs {
			v := os.Getenv(k)
			if len(v) > 0 {
				e.envs[k] = v
			}
		}
	}

	e.once.Do(fn)
}

func (e *EnvOnce) Validate(fn func(k string)) {
	for k, v := range e.envs {
		if strings.TrimSpace(v) == "" {
			fn(k)
		}
	}
}

func (e *EnvOnce) Get(name string) string {
	return e.envs[name]
}

func (e *EnvOnce) Print() {
	for k, v := range e.envs {
		fmt.Printf("[Env] %s=%s\n", k, v)
	}
}

func envNil(k string) {
	log.Fatalf("[Env] %s must not be nil.", k)
}

func init() {

	EtcdStorageEnv.Init()
	EtcdStorageEnv.Print()
	//EtcdStorageEnv.Validate(envNil)

	LdapEnv.Init()
	LdapEnv.Print()
	LdapEnv.Validate(envNil)
}
