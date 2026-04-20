package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxJSONBodySize = 8 << 20

func ParseJSON(r *http.Request, v any) error {
	if r == nil || r.Body == nil {
		return nil
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodySize))
	decoder.UseNumber()

	if err := decoder.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return nil
}
