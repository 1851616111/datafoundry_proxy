package main

import (
	"encoding/json"
	"github.com/golang/glog"
	k8sapi "github.com/openshift/origin/pkg/user/api/v1"
)

func authDF(token string) (*k8sapi.User, error) {
	glog.Infoln(token)
	b, err := httpGet(DF_API_Auth, "Authorization", token)
	if err != nil {
		return nil, err
	}

	user := new(k8sapi.User)
	if err := json.Unmarshal(b, user); err != nil {
		return nil, err
	}

	return user, nil
}

func dfUser(user *k8sapi.User) string {
	return user.Name
}

func getDFUserame(token string) (string, error) {
	glog.Infoln(token)
	user, err := authDF(token)
	if err != nil {
		return "", err
	}
	return dfUser(user), nil
}
