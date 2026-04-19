// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"photo-album/internal/common/response"
	"photo-album/internal/config"
	"photo-album/internal/handler"
	"photo-album/internal/svc"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/usercenter.yaml", "the config file")

func main() {
	flag.Parse()

	c, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("error: config file %s, %s", *configFile, err.Error())
	}

	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		statusCode, body := response.ErrorBody(err)
		return statusCode, body
	})

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func loadConfig(path string) (config.Config, error) {
	var c config.Config
	if err := conf.Load(path, &c); err != nil {
		return c, err
	}

	localPath := localConfigPath(path)
	if _, err := os.Stat(localPath); err == nil {
		var override struct {
			Cos config.CosConfig `json:",optional"`
		}
		if err := conf.Load(localPath, &override); err != nil {
			return c, err
		}
		c.ApplyCosOverride(override.Cos)
	} else if !errors.Is(err, os.ErrNotExist) {
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
