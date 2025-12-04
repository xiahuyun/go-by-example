package main

import (
	"context"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

func main() {
	var cfg logx.LogConf

	logc.MustSetup(cfg)
	defer logc.Close()

	ctx := logx.ContextWithFields(context.Background(), logx.Field("path", "/user/info"))

	logc.Infow(ctx, "hello world")
	logc.Error(ctx, "error log")
}
