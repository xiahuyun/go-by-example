// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"limiter/internal/svc"
)

type ReadyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReadyLogic {
	return &ReadyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReadyLogic) Ready() error {
	time.Sleep(1 * time.Minute)
	return nil
}
