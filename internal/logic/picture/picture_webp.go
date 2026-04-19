package picture

import (
	"encoding/binary"
	"errors"
)

func extractWebPDimensions(header []byte) (int64, int64, error) {
	if len(header) < 30 {
		return 0, 0, errors.New("webp header too short")
	}

	switch string(header[12:16]) {
	case "VP8 ":
		if len(header) < 30 || header[23] != 0x9D || header[24] != 0x01 || header[25] != 0x2A {
			return 0, 0, errors.New("invalid vp8 header")
		}
		width := int64(binary.LittleEndian.Uint16(header[26:28]) & 0x3FFF)
		height := int64(binary.LittleEndian.Uint16(header[28:30]) & 0x3FFF)
		return width, height, nil
	case "VP8L":
		if len(header) < 25 || header[20] != 0x2F {
			return 0, 0, errors.New("invalid vp8l header")
		}
		bits := binary.LittleEndian.Uint32(header[21:25])
		width := int64(bits&0x3FFF) + 1
		height := int64((bits>>14)&0x3FFF) + 1
		return width, height, nil
	case "VP8X":
		width := int64(uint32(header[24])|uint32(header[25])<<8|uint32(header[26])<<16) + 1
		height := int64(uint32(header[27])|uint32(header[28])<<8|uint32(header[29])<<16) + 1
		return width, height, nil
	default:
		return 0, 0, errors.New("unsupported webp chunk")
	}
}
