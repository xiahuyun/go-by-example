package logic

import (
	"context"
	"time"

	"rpc/internal/svc"
	"rpc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(in *proto.PingReq) (*proto.PingResp, error) {
	time.Sleep(50 * time.Millisecond)
	return &proto.PingResp{}, nil
}
