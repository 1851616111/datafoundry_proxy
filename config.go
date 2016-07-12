package main

import (
	"flag"
	"github.com/golang/glog"
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

	DATAFOUNDRY_HOST_ADDR string = "DATAFOUNDRY_HOST_ADDR"
	DATAFOUNDRY_API_ADDR  string = "DATAFOUNDRY_API_ADDR"
	
	ENV_NAME_MYSQL_ADDR       string = "ENV_NAME_MYSQL_ADDR"
	ENV_NAME_MYSQL_PORT       string = "ENV_NAME_MYSQL_PORT"
	ENV_NAME_MYSQL_DATABASE   string = "ENV_NAME_MYSQL_DATABASE"
	ENV_NAME_MYSQL_USER       string = "ENV_NAME_MYSQL_USER"
	ENV_NAME_MYSQL_PASSWORD   string = "ENV_NAME_MYSQL_PASSWORD"
	DONT_UPGRADE_MYSQL_TABLES string = "DONT_UPGRADE_MYSQL_TABLES"
	
	ENV_NAME_KAFKA_ADDR string = "ENV_NAME_KAFKA_ADDR"
	ENV_NAME_KAFKA_PORT string = "ENV_NAME_KAFKA_PORT"
	
	ADMIN_EMAIL_USERNAME string = "ADMIN_EMAIL_USERNAME"
	ADMIN_EMAIL          string = "ADMIN_EMAIL"
	ADMIN_EMAIL_PASSWORD string = "ADMIN_EMAIL_PASSWORD"
	EMAIL_SERVER_HOST    string = "EMAIL_SERVER_HOST"
	EMAIL_SERVER_PORT    string = "EMAIL_SERVER_PORT"
)
const (
	ETCDPrefix      string = "datafoundry.io/"
	ETCDUserPrefix  string = ETCDPrefix + "users/"
	ETCDUserProfile string = ETCDUserPrefix + "%s/profile"
	ETCDUserVerify  string = ETCDPrefix + "verify/%s"

	ETCDOrgsPrefix string = ETCDPrefix + "organizations/%s"
)

var (
	MysqlEnv = &EnvOnce2{EnvOnce: EnvOnce{
		envs: map[string]string{
			ENV_NAME_MYSQL_ADDR:       "",
			ENV_NAME_MYSQL_PORT:       "",
			ENV_NAME_MYSQL_DATABASE:   "",
			ENV_NAME_MYSQL_USER:       "",
			ENV_NAME_MYSQL_PASSWORD:   "",
		},
	}}
	//KafkaEnv = &EnvOnce2{EnvOnce: EnvOnce{
	//	envs: map[string]string{
	//		ENV_NAME_KAFKA_ADDR: "",
	//		ENV_NAME_KAFKA_PORT: "",
	//	},
	//}}
	EmailEnv = &EnvOnce{
		envs: map[string]string{
			ADMIN_EMAIL_USERNAME: "",
			ADMIN_EMAIL:          "",
			ADMIN_EMAIL_PASSWORD: "",
			EMAIL_SERVER_HOST:    "",
			EMAIL_SERVER_PORT:    "",
		},
	}
	
	EtcdStorageEnv = &EnvOnce{
		envs: map[string]string{
			ETCD_HTTP_ADDR: "http://127.0.0.1:2379",
			ETCD_USERNAME:  "",
			ETCD_PASSWORD:  "",
		},
	}
	LdapEnv = &EnvOnce{
		envs: map[string]string{
			LDAP_HOST_ADDR:      "",
			LDAP_ADMIN_USER:     "",
			LDAP_ADMIN_PASSWORD: "",
			LDAP_BASE_DN:        "",
		},
	}
	DataFoundryEnv = &EnvOnce{
		envs: map[string]string{
			DATAFOUNDRY_HOST_ADDR: "dev.dataos.io:8443",
			DATAFOUNDRY_API_ADDR:  "",
		},
	}
	RedisEnv = &EnvOnce{
		envs: map[string]string{"Redis_BackingService_Name": ""},
	}
	DF_HOST     string
	DF_API_Auth string
)

type Env interface {
	Init()
	Validate(func(k string))
	Get(name string) string
	Set(key, value string)
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

func (e *EnvOnce) Set(key, value string) {
	e.envs[key] = value
	return
}

func (e *EnvOnce) Print() {
	for k, v := range e.envs {
		glog.Infof("[Env] %s=%s\n", k, v)
	}
}

type EnvOnce2 struct {
	EnvOnce
}

func (e *EnvOnce2) Init() {
	fn := func() {
		for k := range e.envs {
			v := os.Getenv(os.Getenv(k))
			if len(v) > 0 {
				e.envs[k] = v
			}
		}
	}

	e.once.Do(fn)
}


func envNil(k string) {
	glog.Errorf("[Env] %s must not be nil.", k)
}

func init() {

	flag.Parse()

	MysqlEnv.Init()
	MysqlEnv.Print()
	MysqlEnv.Validate(envNil)
	MysqlEnv.Set(DONT_UPGRADE_MYSQL_TABLES, os.Getenv(DONT_UPGRADE_MYSQL_TABLES))

	//KafkaEnv.Init()
	//KafkaEnv.Print()
	//KafkaEnv.Validate(envNil)
	
	EmailEnv.Init()
	EmailEnv.Print()
	EmailEnv.Validate(envNil)
	
	EtcdStorageEnv.Init()
	EtcdStorageEnv.Print()
	//EtcdStorageEnv.Validate(envNil)

	LdapEnv.Init()
	LdapEnv.Print()
	LdapEnv.Validate(envNil)

	DataFoundryEnv.Init()
	DataFoundryEnv.Set(DATAFOUNDRY_HOST_ADDR, httpsAddrMaker(DataFoundryEnv.Get(DATAFOUNDRY_HOST_ADDR)))
	DataFoundryEnv.Print()

	DF_HOST = DataFoundryEnv.Get(DATAFOUNDRY_HOST_ADDR)
	DF_API_Auth = DF_HOST + "/oapi/v1/users/~"
	glog.Info(DF_HOST, DF_API_Auth)

}
