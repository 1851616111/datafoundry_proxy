package main

type UserInfo struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Aliasname string `json:"nickname,omitempty"`
	Email     string `json:"email"`
}
