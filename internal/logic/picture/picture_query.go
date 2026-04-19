package picture

import (
	"context"
	"fmt"
	"strings"
	"time"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"
	"photo-album/model"
)

const (
	defaultPicturePageNum  = 1
	defaultPicturePageSize = 10
	maxPicturePageSize     = 20
)

func normalizePicturePage(pageNum, pageSize int64) (int64, int64, error) {
	if pageNum <= 0 {
		pageNum = defaultPicturePageNum
	}
	if pageSize <= 0 {
		pageSize = defaultPicturePageSize
	}
	if pageSize > maxPicturePageSize {
		return 0, 0, commonresponse.BadRequest("pageSize 不能超过 20")
	}

	return pageNum, pageSize, nil
}

func buildPublicPictureListWhere(searchText string, tags []string) (string, []any) {
	clauses := []string{"where `isDelete` = 0", "and `reviewStatus` = ?"}
	args := []any{reviewStatusPass}

	searchText = strings.TrimSpace(searchText)
	if searchText != "" {
		pattern := "%" + searchText + "%"
		clauses = append(clauses, "and (`name` like ? or `introduction` like ?)")
		args = append(args, pattern, pattern)
	}

	for _, tag := range normalizeTags(tags) {
		clauses = append(clauses, "and `tags` like ?")
		args = append(args, "%"+tag+"%")
	}

	return strings.Join(clauses, " "), args
}

func buildPictureListWhere(req *types.PictureListRequest, publicOnly bool) (string, []any, error) {
	whereSQL, args := buildBasePictureListWhere(publicOnly)

	if req == nil {
		return whereSQL, args, nil
	}

	if req.Id > 0 {
		whereSQL += " and `id` = ?"
		args = append(args, req.Id)
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		whereSQL += " and `name` like ?"
		args = append(args, "%"+name+"%")
	}
	if introduction := strings.TrimSpace(req.Introduction); introduction != "" {
		whereSQL += " and `introduction` like ?"
		args = append(args, "%"+introduction+"%")
	}
	if category := strings.TrimSpace(req.Category); category != "" {
		whereSQL += " and `category` = ?"
		args = append(args, category)
	}
	for _, tag := range normalizeTags(req.Tags) {
		whereSQL += " and `tags` like ?"
		args = append(args, "%"+tag+"%")
	}
	if req.PicSize > 0 {
		whereSQL += " and `picSize` = ?"
		args = append(args, req.PicSize)
	}
	if req.PicWidth > 0 {
		whereSQL += " and `picWidth` = ?"
		args = append(args, req.PicWidth)
	}
	if req.PicHeight > 0 {
		whereSQL += " and `picHeight` = ?"
		args = append(args, req.PicHeight)
	}
	if req.PicScale > 0 {
		whereSQL += " and `picScale` = ?"
		args = append(args, req.PicScale)
	}
	if picFormat := strings.TrimSpace(req.PicFormat); picFormat != "" {
		whereSQL += " and `picFormat` = ?"
		args = append(args, picFormat)
	}
	if req.UserId > 0 {
		whereSQL += " and `userId` = ?"
		args = append(args, req.UserId)
	}
	if !publicOnly && req.ReviewStatus != nil {
		whereSQL += " and `reviewStatus` = ?"
		args = append(args, *req.ReviewStatus)
	}
	if !publicOnly {
		if reviewMessage := strings.TrimSpace(req.ReviewMessage); reviewMessage != "" {
			whereSQL += " and `reviewMessage` like ?"
			args = append(args, "%"+reviewMessage+"%")
		}
		if req.ReviewerId > 0 {
			whereSQL += " and `reviewerId` = ?"
			args = append(args, req.ReviewerId)
		}
	}
	if searchText := strings.TrimSpace(req.SearchText); searchText != "" {
		pattern := "%" + searchText + "%"
		whereSQL += " and (`name` like ? or `introduction` like ?)"
		args = append(args, pattern, pattern)
	}

	startTime, endTime, err := parseEditTimeRange(req.EditTimeStart, req.EditTimeEnd)
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
		return "where `isDelete` = 0 and `reviewStatus` = ?", []any{reviewStatusPass}
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

func loadRequiredAdmin(ctx context.Context, svcCtx *svc.ServiceContext, authorization string) (*model.User, error) {
	loginUser, err := loadRequiredLoginUser(ctx, svcCtx, authorization)
	if err != nil {
		return nil, err
	}
	if loginUser.UserRole != "admin" {
		return nil, commonresponse.Forbidden("仅管理员可访问")
	}
	return loginUser, nil
}
