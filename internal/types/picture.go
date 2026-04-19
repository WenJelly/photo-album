package types

type PictureGetRequest struct {
	Id int64 `path:"id"`
}

type PictureIDQueryRequest struct {
	Id int64 `json:"id"`
}

type PictureUploadByUrlRequest struct {
	Id           int64    `json:"id,optional"`
	FileUrl      string   `json:"fileUrl"`
	PicName      string   `json:"picName,optional"`
	Introduction string   `json:"introduction,optional"`
	Category     string   `json:"category,optional"`
	Tags         []string `json:"tags,optional"`
}

type PictureUploadRequest struct {
	Id           int64    `json:"id,optional"`
	PicName      string   `json:"picName,optional"`
	Introduction string   `json:"introduction,optional"`
	Category     string   `json:"category,optional"`
	Tags         []string `json:"tags,optional"`
}

type PictureListRequest struct {
	PageNum       int64    `json:"pageNum,optional"`
	PageSize      int64    `json:"pageSize,optional"`
	Id            int64    `json:"id,optional"`
	Name          string   `json:"name,optional"`
	Introduction  string   `json:"introduction,optional"`
	Category      string   `json:"category,optional"`
	Tags          []string `json:"tags,optional"`
	PicSize       int64    `json:"picSize,optional"`
	PicWidth      int64    `json:"picWidth,optional"`
	PicHeight     int64    `json:"picHeight,optional"`
	PicScale      float64  `json:"picScale,optional"`
	PicFormat     string   `json:"picFormat,optional"`
	UserId        int64    `json:"userId,optional"`
	ReviewStatus  *int64   `json:"reviewStatus,optional"`
	ReviewMessage string   `json:"reviewMessage,optional"`
	ReviewerId    int64    `json:"reviewerId,optional"`
	EditTimeStart string   `json:"editTimeStart,optional"`
	EditTimeEnd   string   `json:"editTimeEnd,optional"`
	SearchText    string   `json:"searchText,optional"`
}

type UserSummary struct {
	Id          int64  `json:"id"`
	UserName    string `json:"userName"`
	UserAvatar  string `json:"userAvatar"`
	UserProfile string `json:"userProfile"`
	UserRole    string `json:"userRole"`
}

type PictureResponse struct {
	Id            int64        `json:"id"`
	Url           string       `json:"url"`
	Name          string       `json:"name"`
	Introduction  string       `json:"introduction,optional"`
	Category      string       `json:"category,optional"`
	Tags          []string     `json:"tags,optional"`
	PicSize       int64        `json:"picSize,optional"`
	PicWidth      int64        `json:"picWidth,optional"`
	PicHeight     int64        `json:"picHeight,optional"`
	PicScale      float64      `json:"picScale,optional"`
	PicFormat     string       `json:"picFormat,optional"`
	UserId        int64        `json:"userId"`
	User          *UserSummary `json:"user,optional"`
	CreateTime    string       `json:"createTime"`
	EditTime      string       `json:"editTime"`
	UpdateTime    string       `json:"updateTime"`
	ReviewStatus  int64        `json:"reviewStatus"`
	ReviewMessage string       `json:"reviewMessage,optional"`
	ReviewerId    int64        `json:"reviewerId,optional"`
	ReviewTime    string       `json:"reviewTime,optional"`
	ThumbnailUrl  string       `json:"thumbnailUrl,optional"`
	PicColor      string       `json:"picColor,optional"`
	ViewCount     int64        `json:"viewCount"`
	LikeCount     int64        `json:"likeCount"`
}

type PicturePageResponse struct {
	PageNum  int64              `json:"pageNum"`
	PageSize int64              `json:"pageSize"`
	Total    int64              `json:"total"`
	List     []*PictureResponse `json:"list"`
}

type PictureCarouselResponse struct {
	List []*PictureResponse `json:"list"`
}
