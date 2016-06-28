package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/client"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"log"
	"reflect"
	"time"
)

type Etcd struct {
	client client.KeysAPI
}

func (s *Etcd) Init() (DBStorage, error) {

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
		//s.prepare()
		glog.Info("etcd init successfully.")
		glog.Infoln(cfg.Endpoints, cfg.Username, cfg.Password)
		return s, nil
	}

}

func (s *Etcd) GetValue(key string) (interface{}, error) {
	response, err := s.client.Get(context.Background(), key, nil)
	if err == nil {
		glog.Infof("%+v", response.Node.Value)
		return response.Node.Value, nil
	} else {
		return nil, err
	}

}

func (s *Etcd) SetValue(key string, value interface{}, dir bool) error {
	valuetype := reflect.TypeOf(value).Kind()
	switch valuetype {
	case reflect.String:
		_, err := s.client.Set(context.Background(), key, value.(string), &client.SetOptions{Dir: dir})
		return err
	case reflect.Struct, reflect.Ptr, reflect.Map:
		if b, err := json.Marshal(value); err != nil {
		} else {
			_, err := s.client.Set(context.Background(), key, string(b), &client.SetOptions{Dir: dir})
			return err
		}
	default:
		return errors.New(fmt.Sprintf("unsupport value type %s", valuetype.String()))
	}
	return nil
}

func (s *Etcd) SetValuebyTTL(key string, value interface{}, ttl time.Duration) error {
	valuetype := reflect.TypeOf(value).Kind()
	switch valuetype {
	case reflect.String:
		_, err := s.client.Set(context.Background(), key, value.(string), &client.SetOptions{Dir: false, TTL: ttl})
		return err
	case reflect.Struct, reflect.Ptr, reflect.Map:
		if b, err := json.Marshal(value); err != nil {
		} else {
			_, err := s.client.Set(context.Background(), key, string(b), &client.SetOptions{Dir: false, TTL: ttl})
			return err
		}
	default:
		return errors.New(fmt.Sprintf("unsupport value type %s", valuetype.String()))
	}
	return nil
}

func (s *Etcd) Delete(key string, dir bool) error {
	_, err := s.client.Delete(context.Background(), key, &client.DeleteOptions{Dir: dir})
	if err != nil {
		glog.Error(err)
	}
	return err
}

func (s *Etcd) prepare() {
	glog.Info("prepare....")
	_, err := s.GetValue(ETCDUserPrefix)
	if err != nil {
		glog.Info("err///////", err)
	}
	if checkIfNotFound(err) {
		glog.Info("init etcd structure..")
		_, err = s.client.Set(context.Background(), ETCDUserPrefix, "", &client.SetOptions{Dir: true})
	}
	if err != nil {
		glog.Error(err)
	}
}

func checkIfNotFound(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(client.Error); ok && e.Code == client.ErrorCodeKeyNotFound {
		return true
	}

	return false
}
