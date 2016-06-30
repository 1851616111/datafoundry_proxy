package main

import (
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func ListOrganizations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	http.Error(w, "not implentmented.", http.StatusNotImplemented)
}

func CreateOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	http.Error(w, "not implentmented.", http.StatusNotImplemented)
}

func GetOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	http.Error(w, "not implentmented.", http.StatusNotImplemented)
}

func InviteOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	http.Error(w, "not implentmented.", http.StatusNotImplemented)
}

func AcceptOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	http.Error(w, "not implentmented.", http.StatusNotImplemented)
}

func LeaveOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	http.Error(w, "not implentmented.", http.StatusNotImplemented)
}
func ManageOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	http.Error(w, "not implentmented.", http.StatusNotImplemented)
}
