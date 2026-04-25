package picture

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	commonresponse "photo-album/internal/common/response"
	"photo-album/internal/svc"
	"photo-album/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultPictureCursorPageSize = 30
	maxPictureCursorPageSize     = 60
)

type pictureCursorPayload struct {
	ID int64 `json:"id"`
}

type GetPictureCursorListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPictureCursorListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPictureCursorListLogic {
	return &GetPictureCursorListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPictureCursorListLogic) GetPictureCursorList(req *types.CursorQueryPictureRequest) (*types.PictureCursorPageResponse, error) {
	if req == nil {
		req = &types.CursorQueryPictureRequest{}
	}

	pageSize, err := normalizePictureCursorPage(req.PageSize)
	if err != nil {
		return nil, err
	}

	cursorID, err := decodePictureCursor(req.Cursor)
	if err != nil {
		return nil, commonresponse.BadRequest("cursor is invalid")
	}

	whereSQL, args, err := buildPictureListWhere(cursorRequestToQueryRequest(req), true)
	if err != nil {
		return nil, err
	}

	pictures, err := l.svcCtx.PicturesModel.FindByWhereBeforeID(l.ctx, whereSQL, "`id` desc", cursorID, pageSize+1, args...)
	if err != nil {
		return nil, commonresponse.InternalServerError("query picture cursor list failed")
	}

	hasMore := int64(len(pictures)) > pageSize
	if hasMore {
		pictures = pictures[:pageSize]
	}

	list, err := buildPictureListResponse(l.ctx, l.svcCtx, pictures, true, req.CompressPictureType)
	if err != nil {
		return nil, err
	}

	nextCursor := ""
	if hasMore && len(pictures) > 0 {
		nextCursor, err = encodePictureCursor(pictures[len(pictures)-1].Id)
		if err != nil {
			return nil, commonresponse.InternalServerError("build next cursor failed")
		}
	}

	return &types.PictureCursorPageResponse{
		PageSize:   pageSize,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		List:       list,
	}, nil
}

func normalizePictureCursorPage(pageSize int64) (int64, error) {
	if pageSize <= 0 {
		return defaultPictureCursorPageSize, nil
	}
	if pageSize > maxPictureCursorPageSize {
		return 0, commonresponse.BadRequest("pageSize cannot exceed 60")
	}
	return pageSize, nil
}

func encodePictureCursor(id int64) (string, error) {
	if id <= 0 {
		return "", nil
	}

	data, err := json.Marshal(pictureCursorPayload{ID: id})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodePictureCursor(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}

	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return 0, err
	}

	var payload pictureCursorPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, err
	}
	if payload.ID <= 0 {
		return 0, commonresponse.BadRequest("cursor is invalid")
	}

	return payload.ID, nil
}

func cursorRequestToQueryRequest(req *types.CursorQueryPictureRequest) *types.QueryPictureRequest {
	return &types.QueryPictureRequest{
		Id:                  req.Id,
		Name:                req.Name,
		Category:            req.Category,
		ReviewStatus:        req.ReviewStatus,
		Tags:                req.Tags,
		PicSize:             req.PicSize,
		PicWidth:            req.PicWidth,
		PicHeight:           req.PicHeight,
		PicScale:            req.PicScale,
		PicFormat:           req.PicFormat,
		UserId:              req.UserId,
		ReviewMessage:       req.ReviewMessage,
		ReviewerId:          req.ReviewerId,
		EditTimeStart:       req.EditTimeStart,
		EditTimeEnd:         req.EditTimeEnd,
		SearchText:          req.SearchText,
		CompressPictureType: req.CompressPictureType,
		PageNum:             defaultPicturePageNum,
		PageSize:            req.PageSize,
	}
}
