package main

type UserInfo struct {
	Username    string `json: "username"`
	Password    string `json: "password"`
	Email       string `json: "email"`
	Displayname string `json: "nickname"`
}
