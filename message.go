package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"

	"github.com/asiainfoLDP/datafoundry_proxy/messages"
	//"github.com/asiainfoLDP/datafoundry_proxy/messages/notification"
)

func GetMessages(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	r.Header.Set("User", username)

	messages.GetMyMessages(w, r, params)
}

func GetMessageStat(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	r.Header.Set("User", username)

	messages.GetNotificationStats(w, r, params)
}

func ModifyMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	r.Header.Set("User", username)

	messages.ModifyMessageWithCustomHandler(w, r, params, ModifyMessage_Custom)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	r.Header.Set("User", username)

	messages.DeleteMessage(w, r, params)
}

//===============================================================

const (
	MessageType_SiteNotify = "sitenotify"
	MessageType_AccountMsg = "accountmsg"
	MessageType_Alert      = "alert"
)

const InviteMessage_Hints = "invite,org"            // DON'T CHANGE!
const AcceptOrgIntitation = "accept_org_invitation" // DON'T CHANGE!
type InviteMessage struct {
	OrgID    string `json:"org_id"`
	OrgName  string `json:"org_name"`
	Accepted bool   `json:"accepted"`
}

func SendOrgInviteMessage(receiver, sender, orgId, orgName string) error {
	msg := &InviteMessage{
		OrgID:    orgId,
		OrgName:  orgName,
		Accepted: false,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = messages.CreateInboxMessage(
		MessageType_SiteNotify,
		receiver,
		sender,
		InviteMessage_Hints,
		string(jsonData),
	)

	return err
}

func ModifyMessage_Custom(r *http.Request, params httprouter.Params, m map[string]interface{}) (bool, *messages.Error) {
	action, e := messages.MustStringParamInMap(m, "action", messages.StringParamType_UrlWord)
	if e != nil {
		return false, e
	}

	switch action {
	default:
		return false, nil
	case AcceptOrgIntitation:
		currentUserName, e := messages.MustCurrentUserName(r)
		if e != nil {
			return true, e
		}

		messageid, e := messages.MustIntParamInPath(params, "id")
		if e != nil {
			return true, e
		}

		msg, err := messages.GetMessageByUserAndID(currentUserName, messageid)
		if err != nil {
			return true, messages.GetError2(messages.ErrorCodeGetMessage, err.Error())
		}

		if strings.Index(msg.Hints, InviteMessage_Hints) < 0 {
			return true, messages.GetError2(messages.ErrorCodeInvalidParameters, "not an org invitation message")
		}

		im := &InviteMessage{}

		err = json.Unmarshal([]byte(msg.Raw_data), im)
		if err != nil {
			return true, messages.GetError2(messages.ErrorCodeInvalidParameters, err.Error())
		}

		if im.Accepted {
			return true, messages.GetError2(messages.ErrorCodeInvalidParameters, "already accepted")
		}

		im.Accepted = true

		jsondata, err := json.Marshal(im)
		if err != nil {
			return true, messages.GetError2(messages.ErrorCodeInvalidParameters, err.Error())
		}

		err = messages.ModifyMessageDataByID(messageid, string(jsondata))
		if err != nil {
			return true, messages.GetError2(messages.ErrorCodeInvalidParameters, err.Error())
		}
	}

	return true, nil
}
