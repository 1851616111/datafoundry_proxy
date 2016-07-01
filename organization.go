package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

func ListOrganizations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	resp, _ := json.Marshal(orgs)

	http.Error(w, string(resp), http.StatusOK)
}

func CreateOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	org := new(Orgnazition)
	if err := parseRequestBody(r, org); err != nil {

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	org.CreateTime = time.Now().Format(time.RFC3339)
	resp, _ := json.Marshal(org)
	http.Error(w, string(resp), http.StatusNotImplemented)
}

func GetOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	orgid := ps.ByName("org")
	orgs[0].ID = orgid
	resp, _ := json.Marshal(orgs[0])

	http.Error(w, string(resp), http.StatusOK)
}

/*
func InviteOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	member := new(OrgMember)
	if err := parseRequestBody(r, member); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
*/
func ManageOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	action := ps.ByName("action")
	switch action {
	case "accept":
		http.Error(w, "accept not implentmented.", http.StatusNotImplemented)
	case "leave":

		http.Error(w, "leave not implentmented.", http.StatusNotImplemented)
	case "remove":
		fallthrough
	case "invite":
		fallthrough
	case "privileged":
		member := new(OrgMember)
		if err := parseRequestBody(r, member); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, _ := json.Marshal(member)
		http.Error(w, string(resp), http.StatusOK)
	default:
		http.Error(w, "not supported action", http.StatusBadRequest)
	}
	return

}

//fake response

var orgs = []Orgnazition{
	{
		ID:         "orgs-id128479",
		Name:       "orgname-abc",
		CreateBy:   "san",
		CreateTime: time.Now().Format(time.RFC3339),
		MemberList: memberlist,
	},
	{
		ID:         "orgs-idtrewq",
		Name:       "orgname-vawe",
		CreateBy:   "zxc",
		CreateTime: time.Now().Format(time.RFC3339),
		MemberList: memberlist,
	},
	{
		ID:         "orgs-89876",
		Name:       "orgname-we213",
		CreateBy:   "br",
		CreateTime: time.Now().Format(time.RFC3339),
		MemberList: memberlist,
	},
}

var memberlist = []OrgMember{
	{
		MemberName:   "san",
		IsAdmin:      true,
		PrivilegedBy: "",
		JoinedAt:     time.Now().Format(time.RFC3339),
	},
	{
		MemberName:   "jingxy3",
		IsAdmin:      true,
		PrivilegedBy: "san",
		JoinedAt:     time.Now().Format(time.RFC3339),
	},
	{
		MemberName:   "sx",
		IsAdmin:      true,
		PrivilegedBy: "jingxy3",
		JoinedAt:     time.Now().Format(time.RFC3339),
	},
	{
		MemberName: "jiangtong",
		IsAdmin:    false,
		JoinedAt:   time.Now().Format(time.RFC3339),
	},
	{
		MemberName: "liuxu",
		Status:     OrgMemberStatusInvited,
	},
}
