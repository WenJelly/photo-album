package picture

import (
	"strconv"
	"strings"
)

const blurHashBase83Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz#$%*+,-.:;=?@[]^_{|}~"

func buildSolidBlurHash(hexColor string) string {
	hexColor = strings.TrimSpace(hexColor)
	hexColor = strings.TrimPrefix(hexColor, "#")
	if len(hexColor) != 6 {
		return ""
	}

	color, err := strconv.ParseInt(hexColor, 16, 32)
	if err != nil {
		return ""
	}

	return encodeBlurHashBase83(0, 1) + encodeBlurHashBase83(0, 1) + encodeBlurHashBase83(int(color), 4)
}

func encodeBlurHashBase83(value, length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		digit := value % 83
		result[i] = blurHashBase83Alphabet[digit]
		value /= 83
	}

	return string(result)
}
