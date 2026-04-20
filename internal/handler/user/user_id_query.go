package user

import (
	"photo-album/internal/common/response"
	"photo-album/internal/types"
)

func parseUserIDQuery(raw string) (*types.UserIDQueryRequest, error) {
	id, err := types.ParseSnowflakeID(raw)
	if err != nil || id <= 0 {
		return nil, response.BadRequest("id 必须是正整数")
	}

	return &types.UserIDQueryRequest{Id: id}, nil
}
