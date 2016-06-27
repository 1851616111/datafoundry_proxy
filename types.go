package main

type UserInfo struct {
	Username   string     `json:"username"`
	Password   string     `json:"password,omitempty"`
	Aliasname  string     `json:"nickname,omitempty"`
	Email      string     `json:"email"`
	CreateTime string     `json:"creation_time"`
	Status     UserStatus `json:"status,omitempty"`
}

type UserStatus struct {
	Active      bool   `json:"active"`
	Initialized bool   `json:"initialized"`
	FromLdap    bool   `json:"from_ldap"`
	Phase       string `json:"phase"`
}

const (
	UserStatusActive   string = "active"
	UserStatusInactive string = "inactive"
)

type Password struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
