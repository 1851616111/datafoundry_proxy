package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"

	"github.com/asiainfoLDP/datafoundry_proxy/messages"
)

type mux struct{}

var (
	dbstore DBStorage
)

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	RespError(w, ldpErrorNew(ErrCodeNotFound), http.StatusNotFound)
}

func Forbidden(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	RespError(w, ldpErrorNew(ErrCodeForbidden), http.StatusForbidden)
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
	router := httprouter.New()
	//router.GET("/", Forbidden)
	router.GET("/hello/:name", Hello)

	//users
	//router.GET("/lapi/authenticity_token", Hello)
	router.GET("/lapi/login", Login)
	router.GET("/login", Login)
	router.POST("/lapi/signup", SignUp)
	router.GET("/lapi/user/profile", Profile)
	router.POST("/lapi/password_reset", Hello)          //account_identifier with token.
	router.PUT("/lapi/password_modify", PasswordModify) //account_identifier with token.
	router.POST("/lapi/send_verify_email", SendVerifyMail)
	router.GET("/verify_account/:token", VerifyAccount)

	router.GET("/lapi/inbox", GetMessages)          //get msgs
	router.GET("/lapi/inbox_stat", GetMessageStat)  //get msgs
	router.PUT("/lapi/inbox/:id", ModifyMessage)    //mark msg as read.
	router.DELETE("/lapi/inbox/:id", DeleteMessage) //mark msg as read.
	//organizations
	router.GET("/lapi/orgs", ListOrganizations)
	router.POST("/lapi/orgs", CreateOrganization)
	router.GET("/lapi/orgs/:org", GetOrganization)
	router.DELETE("/lapi/orgs/:org", DeleteOrganization)
	router.PUT("/lapi/orgs/:org/:action", ManageOrganization)
	// router.PUT("/lapi/orgs/:org/accept", JoinOrganization)
	// router.PUT("/lapi/orgs/:org/leave", LeaveOrganization)
	// router.PUT("/lapi/orgs/:org/invite", ManageOrganization)
	// router.PUT("/lapi/orgs/:org/remove", ManageOrganization)     //
	// router.PUT("/lapi/orgs/:org/privileged", ManageOrganization) //
	//action=privileged,remove,

	go messages.Init( /*router, */ MysqlEnv, nil /*KafkaEnv*/, EmailEnv)

	router.NotFound = &mux{}

	log.Fatal(http.ListenAndServe(":9090", router))
}

// 用户：登录，注册，更新，发验证，激活。查询，密码修改，密码找回。
// 消息：
// 组织：创建，删除，成员邀请，成员删除，进组织确认，退出组织，
// 权限管理：权限更改。

func init() {
	dbinit(new(Etcd))
}

func dbinit(driver DBStorage) {
	dbstore, _ = driver.Init()
}
