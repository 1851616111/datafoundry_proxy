package main

type UserInfo struct {
	Username   string     `json:"username"`
	Password   string     `json:"password,omitempty"`
	Aliasname  string     `json:"nickname,omitempty"`
	Email      string     `json:"email"`
	CreateTime string     `json:"creation_time"`
	OrgList    []OrgBrief `json:"orgnazitions,omitempty"`
	Status     UserStatus `json:"status,omitempty"`
	token      string
}
type OrgBrief struct {
	OrgID    string `json:"org_id"`
	OrgName  string `json:"org_name"`
	OrgAlias string `json:"org_alias"`
}

type UserStatus struct {
	Active      bool            `json:"active"`
	Initialized bool            `json:"initialized"`
	FromLdap    bool            `json:"from_ldap"`
	Phase       UserStatusPhase `json:"phase"`
	Quota       UserQuota       `json:"quota"`
}

type UserQuota struct {
	OrgQuota int
	OrgUsed  int
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
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	CreateBy    string         `json:"create_by"`
	CreateTime  string         `json:"creation_time"`
	MemberList  []OrgMember    `json:"members"`
	Status      OrgStatusPhase `json:"status"`
	RoleBinding bool           `json:"rolebinding"`
	Reason      string         `json:"reason,omitempty"`
}

type OrgnazitionList struct {
	Orgnazitions []Orgnazition `json:"orgnazitions"`
}

type OrgMember struct {
	MemberName   string            `json:"member_name"`
	IsAdmin      bool              `json:"privileged"`
	PrivilegedBy string            `json:"privileged_by"`
	JoinedAt     string            `json:"joined_at"`
	Status       MemberStatusPhase `json:"status"`
}

type MemberStatusPhase string

const (
	OrgMemberStatusInvited MemberStatusPhase = "invited"
	OrgMemberStatusjoined  MemberStatusPhase = "joined"
	OrgMemberStatusNone    MemberStatusPhase = "none"
)

type OrgStatusPhase string

const (
	OrgStatusCreated OrgStatusPhase = "created"
	OrgStatusPending OrgStatusPhase = "creating"
	OrgStatusError   OrgStatusPhase = "failed"
)

var ManageActionList = []string{
	"invite",
	"accept",
	"leave",
	"remove",
	"privileged",
}

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status,omitempty"`
	//Data    interface{} `json:"data,omitempty"`
}
