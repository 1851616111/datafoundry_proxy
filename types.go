package main

import (
	"time"
)

type UserInfo struct {
	Username   string     `json:"username"`
	Password   string     `json:"password,omitempty"`
	Aliasname  string     `json:"nickname,omitempty"`
	Email      string     `json:"email"`
	CreateTime time.Time  `json:"creation_time"`
	Status     UserStatus `json:"status,omitempty"`
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type Password struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
