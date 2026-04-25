package picture

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/types"
	"photo-album/model"
)

const (
	defaultPicturePageNum  = 1
	defaultPicturePageSize = 10
	maxPicturePageSize     = 300
)

func normalizePicturePage(pageNum, pageSize int64) (int64, int64, error) {
	if pageNum <= 0 {
		pageNum = defaultPicturePageNum
	}
	if pageSize <= 0 {
		pageSize = defaultPicturePageSize
	}
	if pageSize > maxPicturePageSize {
		return 0, 0, commonresponse.BadRequest("pageSize 不能超过 300")
	}

	return pageNum, pageSize, nil
}

func buildPictureListWhere(req *types.QueryPictureRequest, publicOnly bool) (string, []any, error) {
	if req == nil {
		return buildPictureListWhereInput(pictureListWhereInput{}, publicOnly)
	}

	return buildPictureListWhereInput(
		pictureListWhereInput{
			id:            req.Id,
			name:          req.Name,
			category:      req.Category,
			reviewStatus:  req.ReviewStatus,
			tags:          req.Tags,
			picSize:       req.PicSize,
			picWidth:      req.PicWidth,
			picHeight:     req.PicHeight,
			picScale:      req.PicScale,
			picFormat:     req.PicFormat,
			userID:        req.UserId,
			reviewMessage: req.ReviewMessage,
			reviewerID:    req.ReviewerId,
			editTimeStart: req.EditTimeStart,
			editTimeEnd:   req.EditTimeEnd,
			searchText:    req.SearchText,
		},
		publicOnly,
	)
}

func buildAdminPictureListWhere(req *types.AdminQueryPictureRequest) (string, []any, error) {
	if req == nil {
		return buildPictureListWhereInput(pictureListWhereInput{}, false)
	}
	if req.ReviewStatus < -1 || req.ReviewStatus > reviewStatusReject {
		return "", nil, commonresponse.BadRequest("reviewStatus 只能是 -1、0、1、2")
	}

	return buildPictureListWhereInput(
		pictureListWhereInput{
			id:            req.Id,
			name:          req.Name,
			category:      req.Category,
			reviewStatus:  req.ReviewStatus,
			tags:          req.Tags,
			picSize:       req.PicSize,
			picWidth:      req.PicWidth,
			picHeight:     req.PicHeight,
			picScale:      req.PicScale,
			picFormat:     req.PicFormat,
			userID:        req.UserId,
			reviewMessage: req.ReviewMessage,
			reviewerID:    req.ReviewerId,
			editTimeStart: req.EditTimeStart,
			editTimeEnd:   req.EditTimeEnd,
			searchText:    req.SearchText,
		},
		false,
	)
}

type pictureListWhereInput struct {
	id            string
	name          string
	category      string
	reviewStatus  int64
	tags          []string
	picSize       int64
	picWidth      int64
	picHeight     int64
	picScale      float64
	picFormat     string
	userID        string
	reviewMessage string
	reviewerID    string
	editTimeStart string
	editTimeEnd   string
	searchText    string
}

func buildPictureListWhereInput(input pictureListWhereInput, publicOnly bool) (string, []any, error) {
	whereSQL, args := buildBasePictureListWhere(publicOnly)

	id, err := parseOptionalSnowflakeID(input.id, "id")
	if err != nil {
		return "", nil, err
	}
	if id > 0 {
		whereSQL += " and `id` = ?"
		args = append(args, id)
	}

	if name := strings.TrimSpace(input.name); name != "" {
		whereSQL += " and `name` like ?"
		args = append(args, "%"+name+"%")
	}
	if category := strings.TrimSpace(input.category); category != "" {
		whereSQL += " and `category` = ?"
		args = append(args, category)
	}
	for _, tag := range normalizeTags(input.tags) {
		whereSQL += " and `tags` like ?"
		args = append(args, "%"+tag+"%")
	}
	if input.picSize > 0 {
		whereSQL += " and `picSize` = ?"
		args = append(args, input.picSize)
	}
	if input.picWidth > 0 {
		whereSQL += " and `picWidth` = ?"
		args = append(args, input.picWidth)
	}
	if input.picHeight > 0 {
		whereSQL += " and `picHeight` = ?"
		args = append(args, input.picHeight)
	}
	if input.picScale > 0 {
		whereSQL += " and `picScale` = ?"
		args = append(args, input.picScale)
	}
	if picFormat := strings.TrimSpace(input.picFormat); picFormat != "" {
		whereSQL += " and `picFormat` = ?"
		args = append(args, picFormat)
	}

	userID, err := parseOptionalSnowflakeID(input.userID, "userId")
	if err != nil {
		return "", nil, err
	}
	if userID > 0 {
		whereSQL += " and `userId` = ?"
		args = append(args, userID)
	}

	if !publicOnly {
		if input.reviewStatus >= 0 {
			whereSQL += " and `reviewStatus` = ?"
			args = append(args, input.reviewStatus)
		}
		if reviewMessage := strings.TrimSpace(input.reviewMessage); reviewMessage != "" {
			whereSQL += " and `reviewMessage` like ?"
			args = append(args, "%"+reviewMessage+"%")
		}
		reviewerID, err := parseOptionalSnowflakeID(input.reviewerID, "reviewerId")
		if err != nil {
			return "", nil, err
		}
		if reviewerID > 0 {
			whereSQL += " and `reviewerId` = ?"
			args = append(args, reviewerID)
		}
	}

	if searchText := strings.TrimSpace(input.searchText); searchText != "" {
		pattern := "%" + searchText + "%"
		whereSQL += " and (`name` like ? or `introduction` like ?)"
		args = append(args, pattern, pattern)
	}

	startTime, endTime, err := parseEditTimeRange(input.editTimeStart, input.editTimeEnd)
	if err != nil {
		return "", nil, err
	}
	if !startTime.IsZero() {
		whereSQL += " and `editTime` >= ?"
		args = append(args, startTime)
	}
	if !endTime.IsZero() {
		whereSQL += " and `editTime` <= ?"
		args = append(args, endTime)
	}

	return whereSQL, args, nil
}

func buildBasePictureListWhere(publicOnly bool) (string, []any) {
	if publicOnly {
		return "where `isDelete` = 0 and `reviewStatus` = ?", []any{int64(reviewStatusPass)}
	}

	return "where `isDelete` = 0", nil
}

func parseEditTimeRange(startRaw, endRaw string) (time.Time, time.Time, error) {
	startTime, err := parsePictureTime(startRaw, false)
	if err != nil {
		return time.Time{}, time.Time{}, commonresponse.BadRequest("editTimeStart 格式错误")
	}

	endTime, err := parsePictureTime(endRaw, true)
	if err != nil {
		return time.Time{}, time.Time{}, commonresponse.BadRequest("editTimeEnd 格式错误")
	}

	return startTime, endTime, nil
}

func parsePictureTime(raw string, endOfDay bool) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}

	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, raw, time.Local)
		if err != nil {
			continue
		}
		if layout == "2006-01-02" && endOfDay {
			return parsed.Add(24*time.Hour - time.Second), nil
		}
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("invalid time: %s", raw)
}

func canViewPictureDetail(reviewStatus, ownerUserID int64, loginUser *model.User) bool {
	if reviewStatus == reviewStatusPass {
		return true
	}
	if loginUser == nil {
		return false
	}
	return loginUser.Id == ownerUserID || loginUser.UserRole == "admin"
}

func canManagePicture(ownerUserID int64, loginUser *model.User) bool {
	if loginUser == nil {
		return false
	}

	return loginUser.Id == ownerUserID || loginUser.UserRole == "admin"
}

func parseRequiredSnowflakeID(raw, field string) (int64, error) {
	value, err := parseOptionalSnowflakeID(raw, field)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, commonresponse.BadRequest(field + " 必须是正整数")
	}
	return value, nil
}

func parseOptionalSnowflakeID(raw, field string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, commonresponse.BadRequest(field + " 必须是正整数")
	}

	return id, nil
}
