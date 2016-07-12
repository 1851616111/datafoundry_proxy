package main

import (
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	
	"github.com/asiainfoLDP/datafoundry_proxy/messages"
)

func GetMessages(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	r.Header.Set("User", username)
	
	messages.GetMyMessages(w, r, params)
}

func ModifyMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	r.Header.Set("User", username)
	
	messages.ModifyMessage(w, r, params)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	r.Header.Set("User", username)
	
	messages.DeleteMessage(w, r, params)
}

//===============================================================

type InviteMessage struct {
	OrgID    string  `json:"org_id"`
	OrgName  string  `json:"org_name"`
}

func SendOrgInviteMessage(receiver, sender, orgId, orgName string) error {
	msg := &InviteMessage {
		OrgID:   orgId,
		OrgName: orgName,
	}
	
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	_, err = messages.CreateInboxMessage(
		messages.Message_SiteNotify, 
		receiver, 
		sender, 
		string(jsonData),
	)
	
	return err
}
