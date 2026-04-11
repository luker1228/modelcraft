package profile

// GetMyUserProfileCommand 查询当前用户及其资料的命令。
type GetMyUserProfileCommand struct {
	OrgName string
	UserID  string
}

// UserProfileView 聚合返回 user + profile 视图。
type UserProfileView struct {
	User    UserView
	Profile ProfileView
}

// UserView 是 user 信息的应用层视图。
type UserView struct {
	ID        string
	Phone     string
	UserName  string
	Status    string
	CreatedAt string
	UpdatedAt string
}

// ProfileView 是 profile 信息的应用层视图。
type ProfileView struct {
	ID        string
	UserID    string
	Nickname  string
	AvatarURL *string
	Bio       *string
	CreatedAt string
	UpdatedAt string
}

// UpdateMyProfileCommand 更新当前用户资料。
type UpdateMyProfileCommand struct {
	OrgName   string
	UserID    string
	Nickname  *string
	AvatarURL *string
	Bio       *string
}
