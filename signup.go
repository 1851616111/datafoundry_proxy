package main

import (
	"github.com/julienschmidt/httprouter"

	"net/http"
)

func SignUp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//fmt.Println("method:",r.Method)
	//fmt.Println("scheme", r.URL.Scheme)

	r.ParseForm()
}
