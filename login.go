package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
)

var apiserver = "https://dev.dataos.io:8443/oauth/authorize?client_id=openshift-challenging-client&response_type=token"
var oauthurl = "https://datafoundry-oauth.app.dataos.io/v1/repos/gitlab/login"
var gitlaburl = "https://code.dataos.io"

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//fmt.Println("method:",r.Method)
	//fmt.Println("scheme", r.URL.Scheme)

	r.ParseForm()
	switch r.Method {
	case "GET":
		auth := r.Header.Get("Authorization")
		if len(auth) > 0 {
			glog.Infoln(auth)
			token, status := token_proxy(auth)
			if len(token) > 0 {
				glog.Infoln(token)

				resphttp(w, http.StatusOK, []byte(token))
			} else {
				glog.Infoln("error from server, code:", status)
				resphttp(w, status, nil)
			}

		} else {
			RespError(w, ldpErrorNew(ErrCodeUnauthorized), http.StatusUnauthorized)
		}
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.WriteHeader(http.StatusNoContent)
	default:
		RespError(w, ldpErrorNew(ErrCodeMethodNotAllowed), http.StatusMethodNotAllowed)
	}

}

func logingitlab(basic string, auth map[string]string) {
	//log.Println(basic, auth)
	var bearer, posturl string
	if len(auth["token_type"]) == 0 || len(auth["access_token"]) == 0 {
		glog.Infoln(auth, "doesn't contain a complete token")
	} else {
		bearer = auth["token_type"] + " " + auth["access_token"]
	}

	b64auth := strings.Split(basic, " ")
	if len(b64auth) != 2 {
		glog.Infoln("basic string error.")
		return
	} else {
		payload, _ := base64.StdEncoding.DecodeString(b64auth[1])
		pair := strings.Split(string(payload), ":")
		if len(pair) != 2 {
			glog.Infoln(pair, "doesn't contain a username or password.")
			return
		} else {
			posturl = fmt.Sprintf("%s?host=%s&username=%s&password=%s", oauthurl, gitlaburl, pair[0], pair[1])
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("POST", posturl, nil)
	req.Header.Set("Authorization", bearer)
	//log.Println(req.Header, bearer)

	resp, err := client.Do(req)
	if err != nil {
		glog.Error(err)
	} else {
		glog.Infoln(req.Host, req.Method, req.URL.RequestURI(), req.Proto, resp.StatusCode)
	}
	return
}

func token_proxy(auth string) (token string, status int) {
	//fmt.Println("prepear to get token from", url, "with", auth)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//RoundTrip:       roundTrip,
	}

	var DefaultTransport http.RoundTripper = tr

	req, _ := http.NewRequest("HEAD", apiserver, nil)
	req.Header.Set("Authorization", auth)

	resp, err := DefaultTransport.RoundTrip(req)
	//defer resp.Body.Close()

	//resp, err := client.Do(req)
	if err != nil {
		glog.Error(err)
		return "", http.StatusInternalServerError
	} else {
		url, err := resp.Location()
		if err == nil {
			//fmt.Println("resp", url.Fragment)
			m := strings.Split(url.Fragment, "&")
			n := proc(m)
			r, _ := json.Marshal(n)
			go logingitlab(auth, n)
			checkIfInitProject(n)
			return string(r), resp.StatusCode
		}
	}
	return "", resp.StatusCode
}

func proc(s []string) (m map[string]string) {
	m = map[string]string{}
	for _, v := range s {
		n := strings.Split(v, "=")
		m[n[0]] = n[1]
	}
	return
}

func resphttp(rw http.ResponseWriter, code int, body []byte) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Authorization")
	rw.WriteHeader(code)
	rw.Write(body)
}

func init() {
	api := os.Getenv("DATAFOUNDRY_APISERVER_ADDR")
	if len(api) > 0 {
		apiserver = "https://" + api + "/oauth/authorize?client_id=openshift-challenging-client&response_type=token"
	}

	oauthserver := os.Getenv("DATAFOUNDRY_OAUTH_URL")
	if len(oauthserver) > 0 {
		oauthurl = oauthserver
	}

	gitserver := os.Getenv("DATAFOUNDRY_GIT_ADDR")
	if len(gitserver) > 0 {
		gitlaburl = gitserver
	}
	glog.Infoln("apiserver", apiserver)
	glog.Infoln("oauthurl", oauthurl)
	glog.Infoln("gitlaburl", gitlaburl)
}

func checkIfInitProject(auth map[string]string) {
	var token string
	if len(auth["token_type"]) == 0 || len(auth["access_token"]) == 0 {
		glog.Infoln(auth, "doesn't contain a complete token")
	} else {
		token = auth["token_type"] + " " + auth["access_token"]
	}

	if username, err := getDFUserame(token); err != nil {
		glog.Error(err)
	} else {
		if profile, err := getProfile(username); err != nil {
			glog.Error(err)
		} else {
			user := new(UserInfo)
			if err = json.Unmarshal([]byte(profile.(string)), user); err != nil {
				glog.Error(err)
			} else {
				if user.Status.Initialized {
					return
				} else {
					if err = user.InitUserProject(token); err != nil {
						glog.Error(err)
						return
					}
				}
			}
		}
	}
}
