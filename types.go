package main

type UserInfo struct {
	Username   string     `json:"username"`
	Password   string     `json:"password,omitempty"`
	Aliasname  string     `json:"nickname,omitempty"`
	Email      string     `json:"email"`
	CreateTime string     `json:"creation_time"`
	OrgList    []string   `json:"orgnazitions,omitempty"`
	Status     UserStatus `json:"status,omitempty"`
}

type UserStatus struct {
	Active      bool            `json:"active"`
	Initialized bool            `json:"initialized"`
	FromLdap    bool            `json:"from_ldap"`
	Phase       UserStatusPhase `json:"phase"`
}

type UserStatusPhase string

const (
	UserStatusActive   UserStatusPhase = "active"
	UserStatusInactive UserStatusPhase = "inactive"
)

type Password struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type Orgnazition struct {
	Name        string      `json:"name"`
	DisplayName string      `json:"display_name"`
	CreateTime  string      `json:"creation_time"`
	MemberList  []OrgMember `json:"members"`
}

type OrgMember struct {
	MemberName   string `json:"member_name"`
	IsAdmin      bool   `json:"privileged"`
	PrivilegedBy string `json:"privileged_by"`
	JoinedAt     string `json:"joined_at"`
}
