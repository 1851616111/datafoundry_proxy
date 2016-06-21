package main

import (
	"github.com/coreos/etcd/client"
	"log"
	"time"
)

type Etcd struct {
	client client.KeysAPI
}

func (s *Etcd) Init() (DBStorage, error) {
	EtcdStorageEnv.Init()
	EtcdStorageEnv.Print()
	//EtcdStorageEnv.Validate(envNil)

	//初始化etcd客户端
	cfg := client.Config{
		Endpoints: []string{httpAddrMaker(EtcdStorageEnv.Get(ETCD_HTTP_ADDR))},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
		Username:                EtcdStorageEnv.Get(ETCD_USERNAME),
		Password:                EtcdStorageEnv.Get(ETCD_PASSWORD),
	}

	if c, err := client.New(cfg); err != nil {
		log.Fatal("Can not init ectd client", err)
		return nil, err
	} else {
		s.client = client.NewKeysAPI(c)
		return s, nil
	}

}

func (s *Etcd) GetValue(key string) (string, error) {
	return "", nil
}

func (s *Etcd) SetValue(key string, value interface{}) error {
	return nil
}
