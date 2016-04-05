package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type waylandMux struct {
}

/*
func NotFoundw(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Hello you are visiting path %s, but it doesn't exist.\n",r.URL.Path)
}
func (p *waylandMux)NotFoundHandler()http.Handler{
    return http.HandlerFunc(NotFoundw)
}
*/
func (p *waylandMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Println("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	switch r.URL.Path {
	case "/login":
		login(w, r)
	default:
		http.Error(w, "", http.StatusForbidden)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("method:",r.Method)
	//fmt.Println("scheme", r.URL.Scheme)

	r.ParseForm()
	switch r.Method {
	case "GET":
		auth := r.Header.Get("Authorization")
		if len(auth) > 0 {
			log.Println(auth)
			token := token_proxy(auth)
			if len(token) > 0 {
				log.Println(token)
				resphttp(w, http.StatusOK, []byte(token))
			}

		} else {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowd.", http.StatusMethodNotAllowed)
	}

}

var url = "https://54.222.199.235:8443/oauth/authorize?client_id=openshift-challenging-client&response_type=token"

func token_proxy(auth string) (token string) {
	//fmt.Println("prepear to get token from", url, "with", auth)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//RoundTrip:       roundTrip,
	}

	var DefaultTransport http.RoundTripper = tr

	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("Authorization", auth)

	resp, err := DefaultTransport.RoundTrip(req)
	//defer resp.Body.Close()

	//resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	} else {
		url, err := resp.Location()
		if err == nil {
			//fmt.Println("resp", url.Fragment)
			m := strings.Split(url.Fragment, "&")
			n := proc(m)
			r, _ := json.Marshal(n)
			return string(r)
		}
	}
	return
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

func main() {
	mux := &waylandMux{}

	port := ":9090"
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Println("Hello, world!")
}

func init() {
	apiserver := os.Getenv("DATAFOUNDRY_APISERVER_ADDR")
	if len(apiserver) > 0 {
		url = "https://" + apiserver + "/oauth/authorize?client_id=openshift-challenging-client&response_type=token"
	}
	log.Println(url)
}
