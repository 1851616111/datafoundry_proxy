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

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if orgs, err := userinfo.ListOrgs(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		resp, _ := json.Marshal(orgs)
		http.Error(w, string(resp), http.StatusOK)
	}
	return

}

func CreateOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	org := new(Orgnazition)
	if err := parseRequestBody(r, org); err != nil {

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if userinfo.CheckIfOrgExist(org.Name) {
		http.Error(w, "organization name already exist.", http.StatusBadRequest)
		return
	}

	if org, err = userinfo.CreateOrg(org); err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		if resp, err := json.Marshal(org); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, string(resp), http.StatusOK)
		}
		return
	}

}

func GetOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID := ps.ByName("org")
	if !userinfo.CheckIfOrgExistByID(orgID) {
		http.Error(w, "no such organization", http.StatusNotFound)
		return
	}

	orgnazition := new(Orgnazition)
	orgnazition.ID = orgID

	if orgnazition, err = orgnazition.Get(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		resp, _ := json.Marshal(orgnazition)
		http.Error(w, string(resp), http.StatusOK)
	}
	return
}

func ManageOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	// var username string
	// var err error

	// if username, err = authedIdentities(r); err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

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

func (o *Orgnazition) Get() (org *Orgnazition, err error) {
	obj, err := dbstore.GetValue(etcdOrgPath(o.ID))
	if err != nil {
		return nil, err
	}
	org = new(Orgnazition)
	if err = json.Unmarshal([]byte(obj.(string)), org); err != nil {
		glog.Error(err)
		return nil, err
	}
	return org, nil

}

func (o *Orgnazition) AddMemeber(member *OrgMember) *Orgnazition { return nil }

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
