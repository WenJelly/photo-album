package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type SnowflakeID int64

func NewSnowflakeID(value int64) SnowflakeID {
	return SnowflakeID(value)
}

func ParseSnowflakeID(raw string) (SnowflakeID, error) {
	var id SnowflakeID
	if err := id.parse(raw); err != nil {
		return 0, err
	}

	return id, nil
}

func (id SnowflakeID) Int64() int64 {
	return int64(id)
}

func (id SnowflakeID) String() string {
	if id <= 0 {
		return ""
	}

	return strconv.FormatInt(int64(id), 10)
}

func (id SnowflakeID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *SnowflakeID) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*id = 0
		return nil
	}

	if data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}

		return id.parse(raw)
	}

	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err != nil {
		return err
	}

	return id.parse(number.String())
}

func (id *SnowflakeID) UnmarshalText(text []byte) error {
	return id.parse(string(text))
}

func (id *SnowflakeID) parse(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		*id = 0
		return nil
	}

	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid snowflake id: %w", err)
	}
	if parsed < 0 {
		return fmt.Errorf("invalid snowflake id: must be non-negative")
	}

	*id = SnowflakeID(parsed)
	return nil
}
