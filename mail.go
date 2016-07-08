package main

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/jordan-wright/email"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

var AdminName = "DataFoundry Team"
var AdminEmailUser = ""
var AdminEmailPass = ""
var EmailServerHost = ""
var EmailServerPort = 0

func init() {
	name := os.Getenv("ADMIN_EMAIL_USERNAME")
	if name != "" {
		AdminName = name
	}

	email := os.Getenv("ADMIN_EMAIL")
	if email != "" {
		AdminEmailUser = email
	}

	pass := os.Getenv("ADMIN_EMAIL_PASSWORD")
	if pass != "" {
		AdminEmailPass = pass
	}

	host := os.Getenv("EMAIL_SERVER_HOST")
	if host != "" {
		EmailServerHost = host
	}

	port_str := os.Getenv("EMAIL_SERVER_PORT")
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

var bccEmail []string = []string{
	"chaizs@asiainfo.com",
	"jingxy3@asiainfo.com",
	"xueys@maitewang.com",
	"jiangtong@asiainfo.com",
}

var Subject string = "Welcome to Datafoundry"
var Message1 string = `Hello %s, <br />please click <a href="%s">link</a> to verify your account, the activation link will be expire after 24 hours.`

var Message string = `<body style="background-color: #e6e9f2">
    <div class="sendemailbox">
        <div class="sendemail" style="margin:auto; background-color: white; width: 540px;">
            <div style="padding: 30px 30px">
                <div class="headLogo">
                    <img src="https://lab.dataos.io/pub/img/mail_banner.png" >
                </div>
                <div class="content" style="">
                    <p style="font-size: 16px; margin: 50px 0 15px 0px;color: #5a6378;">亲爱的 %s，感谢您注册<a style="font-size: 16px;color: #000000;">&nbsp;铸数工坊&nbsp;DataFoundry</a>。</p>
                    <p style="font-size: 16px;color: #5a6378; margin-bottom: 20px">请点击按钮激活您的账号。</p>
                    <div class="button" style="height: 100px;">
                        <a href="%s" ID="activation" type="button" style="height: 40px; width: 160px; font-size: 18px; background-color: #f6a540; border: 1px solid #f6a540; border-radius: 2px; margin-bottom: 60px; color: white; padding: 10px 20px; text-decoration:none">立即验证邮箱</a>
                    </div>
                    <div class="submes" style="font-size: 14px; color: #5a6378;">
                        <p>如果按钮无法点击，请将下面的链接复制到浏览器地址栏中打开：</p>
                        <p class="address" style="margin-top: 15px">%s</p>
                        <p style="margin-bottom: 30px">请您在 24 小时内激活。</p>
                    </div>
                </div>
            </div>

            <div class="buttom" style=" width:540px;">
                <div style="padding: 30px 30px; background-color: #f7f8fb;">
                    <img src="https://lab.dataos.io/pub/img/mail_qrcode.png">
                    <div style="padding-top: 0px; width:50%; color: #000000;display: inline-block; margin-left: 20px; font-size: 12px">
                        <p style="; ">扫一扫</p>
                        <p style="margin-top: 10px">了解最新产品和咨询</p>
                        <p style="margin-top: 20px; color: #ef9033">铸数工坊公众号</p>
                    </div>
                </div>
            </div>
        </div>
        <div style="margin:0 auto; padding-top: 30px; padding-bottom: 35px; width: 540px; position:relative">
            <div class="line">
            </div>
            <img id="footer" src="https://lab.dataos.io/pub/img/mail_logo_small.png" style="background-color: #e6e9f2; position: absolute; top: 15px; left:45%; height: 30px; width:30px; z-index: 5; padding: 0px 10px;">
        </div>
    </div>

</body>
<style>
    #activation :hover {
        background-color: #e5993a;
    }
    #activation :active {
        background-color: #f8b551;
    }
    .line {
        width:100%;
        border-bottom: 1px solid #c9d0e2;
    }
</style>`
