package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func VerifyAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	r.ParseForm()
	token := ps.ByName("token")
	if len(token) == 0 {
		RespError(w, ldpErrorNew(ErrCodeInvalidToken), http.StatusBadRequest)
		return
	}
	user, err := dbstore.GetValue(etcdGeneratePath(ETCDUserVerify, token))
	if err != nil {
		if checkIfNotFound(err) {
			glog.Errorf("token %s not exist.", token)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(VerifyFail))

		} else {
			glog.Error(err)
			RespError(w, err, http.StatusInternalServerError)
		}
		return
	}

	if err = activeAccount(user.(string), token); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(VerifySuccess))
	}
}

func activeAccount(user, token string) error {
	profile, err := dbstore.GetValue(etcdProfilePath(user))
	if err != nil {
		glog.Error(err)
		return err
	}
	userinfo := new(UserInfo)
	if err = json.Unmarshal([]byte(profile.(string)), userinfo); err != nil {
		glog.Error(err)
		return err
	} else {
		userinfo.Status.Active = true
	}
	glog.Warning("TODO: INIT USER, CREATE NEW PROJECT.")

	if err = dbstore.SetValue(etcdProfilePath(userinfo.Username), userinfo, false); err != nil {
		glog.Error(err)
		return err
	} else {
		return dbstore.Delete(etcdGeneratePath(ETCDUserVerify, token), false)
	}

}

var VerifySuccess string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>激活成功</title>
</head>
<body style="background-color: #e6e9f2;">
    <div class="box" style="width:460px; height: 350px;background: white; margin:0 auto; position: relative">
        <div class="top" style="height: 270px; width:460px; text-align: center">
            <img src="/pub/img/mail_logo_color.png" style="padding: 30px 0;">
            <p style="font-size: 16px; color: #000000; margin-left: 5px">激活成功</p>
            <p style="padding: 20px 0 50px 0; font-size: 14px; color: #5a6378;">您的帐号已完成激活。</p>
        </div>
        <div class="bottom">
            <a  href="/#/login" class="login" type="button" style="text-decoration:none">登录</a>
        </div>
    </div>
    <div style="margin:0 auto; padding-top: 30px; padding-bottom: 30px; width: 460px; position:relative">
        <div class="line">
        </div>
        <img id="footer" src="/pub/img/mail_logo_small.png">
    </div>
</body>
<style>
    .login{
        height:40px;
        width: 120px;
        font-size: 18px;
        background-color: #f6a540;
        border: 1px solid #f6a540;
        border-radius: 2px;
        color: white;
        padding: 10px 50px 10px 50px;
    }
    .line {
        width: 100%;
        border-bottom: 1px solid #c9d0e2;
    }
    .bottom{
        padding-top: 30px;
        padding-bottom: 32px;
        background-color: #f7f8fb;
        text-align: center
    }
    #footer{
        background-color: #e6e9f2;
        position: absolute;
        top: 15px;
        left:45%;
        height: 30px;
        width:30px;
        z-index: 5;
        padding: 0px 10px;
    }
</style>
</html>`

var VerifyFail string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>激活失败</title>
</head>
<body style="background-color: #e6e9f2;">
<div class="box" style="width:460px; height: 350px;background: white; margin:0 auto; position: relative">
    <div class="top" style="height: 270px; width:460px; text-align: center">
        <img src="/pub/img/mail_logo_grey.png" style="padding: 30px 0;">
        <p style="font-size: 16px; color: #000000; margin-left: 5px">激活失败</p>
        <p style="margin-left:10px; padding: 7px 0 50px 0; font-size: 14px; color: #5a6378; ">链接超过24小时已失效, <br>请登录系统重新发送验证邮件。</p>
    </div>
    <div class="bottom">
        <a class="login" type="button" href="/#/login" style="text-decoration:none">登录</a>
    </div>
</div>
<div style="margin:0 auto; padding-top: 25px; padding-bottom: 30px; width: 460px; position:relative">
    <div class="line">
    </div>
    <img id="footer" src="/pub/img/mail_logo_small.png">
</div>
</body>
<style>
    .login{
        height:40px;
        width: 120px;
        font-size: 18px;
        background-color: #f6a540;
        border: 1px solid #f6a540;
        border-radius: 2px;
        color: white;
        padding: 10px 50px 10px 50px;
    }
    .bottom {
        padding-top: 30px;
        padding-bottom: 32px;
        background-color: #f7f8fb;
        text-align: center
    }
    .line {
        width: 100%;
        border-bottom: 1px solid #c9d0e2;
    }
    #footer {
        background-color: #e6e9f2;
        position: absolute;
        top: 15px;
        left:45%;
        height: 30px;
        width:30px;
        z-index: 5;
        padding: 0px 10px;
    }
</style>
</html>
`
