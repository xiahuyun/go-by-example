// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"limiter/internal/svc"
	"limiter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserinfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserinfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserinfoLogic {
	return &UserinfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserinfoLogic) Userinfo(req *types.UserInfoRequest) (resp *types.UserInfoResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
