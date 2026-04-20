package picture

import (
	"photo-album/internal/common/response"
	"photo-album/internal/types"
)

func parsePictureIDQuery(raw string) (*types.PictureIDQueryRequest, error) {
	id, err := types.ParseSnowflakeID(raw)
	if err != nil || id <= 0 {
		return nil, response.BadRequest("id 必须是正整数")
	}

	return &types.PictureIDQueryRequest{Id: id}, nil
}
