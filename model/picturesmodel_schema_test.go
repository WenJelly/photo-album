package model

import (
	"strings"
	"testing"
)

func TestPicturesRuntimeRowsDoNotRequireBlurHashColumn(t *testing.T) {
	if strings.Contains(picturesRows, "`blurHash`") {
		t.Fatal("picture queries must not require optional blurHash column before migration")
	}
}
