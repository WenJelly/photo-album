package types

type DetailUserRequest struct {
	Id        string `json:"id,optional"`
	UserEmail string `json:"userEmail,optional"`
}

type LoginRequest struct {
	UserEmail    string `json:"userEmail" validate:"required"`
	UserPassword string `json:"userPassword" validate:"required"`
}

type LoginResponse struct {
	Token       string `json:"token"`
	Id          string `json:"id"`
	UserEmail   string `json:"userEmail"`
	UserName    string `json:"userName"`
	UserAvatar  string `json:"userAvatar"`
	UserProfile string `json:"userProfile"`
	UserRole    string `json:"userRole"`
	CreateTime  string `json:"createTime"`
	UpdateTime  string `json:"updateTime"`
}

type RegisterRequest struct {
	UserEmail         string `json:"userEmail" validate:"required"`
	UserPassword      string `json:"userPassword" validate:"required"`
	UserCheckPassword string `json:"userCheckPassword" validate:"required"`
}

type RegisterResponse struct {
	Id string `json:"id"`
}

type UpdateUserRequest struct {
	Id           string `json:"id"`
	UserName     string `json:"userName,optional"`
	UserEmail    string `json:"userEmail,optional"`
	UserPassword string `json:"userPassword,optional"`
	UserAvatar   string `json:"userAvatar,optional"`
	UserProfile  string `json:"userProfile,optional"`
}

type AdminUpdateUserRequest struct {
	Id           string `json:"id"`
	UserName     string `json:"userName,optional"`
	UserEmail    string `json:"userEmail,optional"`
	UserPassword string `json:"userPassword,optional"`
	UserAvatar   string `json:"userAvatar,optional"`
	UserProfile  string `json:"userProfile,optional"`
	UserRole     string `json:"userRole,optional"`
}

type DetailUserResponse struct {
	Id                   string `json:"id"`
	UserName             string `json:"userName"`
	UserEmail            string `json:"userEmail"`
	UserAvatar           string `json:"userAvatar"`
	UserProfile          string `json:"userProfile"`
	UserRole             string `json:"userRole"`
	CreateTime           string `json:"createTime"`
	UpdateTime           string `json:"updateTime"`
	PictureCount         int64  `json:"pictureCount"`
	ApprovedPictureCount int64  `json:"approvedPictureCount"`
	PendingPictureCount  int64  `json:"pendingPictureCount"`
	RejectedPictureCount int64  `json:"rejectedPictureCount"`
}
