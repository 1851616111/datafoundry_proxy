package main

type UserInfo struct {
	Username  string     `json:"username"`
	Password  string     `json:"password,omitempty"`
	Aliasname string     `json:"nickname,omitempty"`
	Email     string     `json:"email"`
	Status    UserStatus `json:"status,omitempty"`
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)
