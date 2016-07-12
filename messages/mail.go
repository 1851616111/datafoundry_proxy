package messages

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/jordan-wright/email"
	"net/mail"
	"net/smtp"
	//"os"
	"strconv"
	"strings"
)

var AdminName = "DataFoundry Team"
var AdminEmailUser = ""
var AdminEmailPass = ""
var EmailServerHost = ""
var EmailServerPort = 0

func initMail() {
	//name := os.Getenv("ADMIN_EMAIL_USERNAME")
	name := emailEnv.Get("ADMIN_EMAIL_USERNAME")
	if name != "" {
		AdminName = name
	}

	//email := os.Getenv("ADMIN_EMAIL")
	email := emailEnv.Get("ADMIN_EMAIL")
	if email != "" {
		AdminEmailUser = email
	}

	//pass := os.Getenv("ADMIN_EMAIL_PASSWORD")
	pass := emailEnv.Get("ADMIN_EMAIL_PASSWORD")
	if pass != "" {
		AdminEmailPass = pass
	}

	//host := os.Getenv("EMAIL_SERVER_HOST")
	host := emailEnv.Get("EMAIL_SERVER_HOST")
	if host != "" {
		EmailServerHost = host
	}

	//port_str := os.Getenv("EMAIL_SERVER_PORT")
	port_str := emailEnv.Get("EMAIL_SERVER_PORT")
	if port_str != "" {
		port, err := strconv.ParseInt(port_str, 10, 64)
		if err != nil {
			glog.Error(err)
		} else {
			EmailServerPort = int(port)
		}
	}

	glog.Infof("Mail server and account: @%s:%d, %s(%s)", EmailServerHost, EmailServerPort, AdminName, AdminEmailUser)
}

//================================================================
//
//================================================================

var FromAddress string

func getFromAddress() string {
	if FromAddress == "" {
		FromAddress = (&mail.Address{Name: AdminName, Address: AdminEmailUser}).String()
	}

	return FromAddress
}

var EmailServerAddr string

func getEmailServerAddr() string {
	if EmailServerAddr == "" {
		EmailServerAddr = fmt.Sprintf("%s:%d", EmailServerHost, EmailServerPort)
	}

	return EmailServerAddr
}

func EmailsString2EmailList(emailsString string) []string {
	emails := strings.Split(emailsString, ",")
	num := len(emails)
	index := 0
	for i := 0; i < num; i++ {
		email := strings.TrimSpace(emails[i])
		if len(email) > 0 { // todo: use common.ValidateEmail ?
			emails[index] = email
			index++
		}
	}

	return emails[:index]
}

func SendMail(to, cc, bcc []string, subject, message string, isHtml bool) error {
	if len(to) == 0 && len(cc) == 0 && len(bcc) == 0 {
		return errors.New("There must be on non-blank list in to, cc and bcc")
	}

	e := email.NewEmail()
	tosend := false
	if len(to) > 0 {
		e.To = to
		tosend = true
	}
	if len(cc) > 0 {
		e.Cc = cc
		tosend = true
	}
	if len(bcc) > 0 {
		e.Bcc = bcc
		tosend = true
	}
	if tosend == false {
		return errors.New("to/cc/bcc are all blank")
	}
	e.From = getFromAddress()
	e.Subject = subject
	if isHtml {
		e.HTML = []byte(message)
	} else {
		e.Text = []byte(message)
	}

	return e.Send(getEmailServerAddr(), smtp.PlainAuth("", AdminEmailUser, AdminEmailPass, EmailServerHost))
}
