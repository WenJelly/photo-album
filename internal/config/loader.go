package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
)

const DefaultConfigPath = "etc/photo-album.yaml"

func Load(path string) (Config, error) {
	var c Config

	path = strings.TrimSpace(path)
	if path == "" {
		path = DefaultConfigPath
	}

	if err := conf.Load(path, &c); err != nil {
		return c, err
	}

	localPath := localConfigPath(path)
	if _, err := os.Stat(localPath); err == nil {
		var override struct {
			Cos CosConfig `json:",optional"`
		}
		if err := conf.Load(localPath, &override); err != nil {
			return c, err
		}
		c.ApplyCosOverride(override.Cos)
	} else if !os.IsNotExist(err) {
		return c, err
	}

	c.Normalize()
	return c, nil
}

func localConfigPath(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path + ".local"
	}

	return strings.TrimSuffix(path, ext) + ".local" + ext
}
