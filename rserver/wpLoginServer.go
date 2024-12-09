package rserver

import (
	"connectrpc.com/connect"
	"context"
	"muskex/gen/mproto"
	"muskex/mmlogin"
	"muskex/mmlogin/application/auth"
)

// tron 钱包插件登陆
type LoginServer struct {
}

//	func init() {
//		mmlogin.InitMMLogin()
//	}
//
// 插件获取登陆授权码，后续app把授权码签名后，会把签名提交回来。
func (ls LoginServer) WPChallenge(ctx context.Context, req *connect.Request[mproto.ChallengeRequest]) (*connect.Response[mproto.ChallengeResponse], error) {
	chCode, err := mmlogin.Apps.Auth.Challenge(ctx, auth.NewChallengeInput(req.Msg.Address))
	if err != nil {
		return nil, err
	}
	res := connect.NewResponse(&mproto.ChallengeResponse{Code: chCode.Challenge})
	return res, nil
}

// 登陆： app把授权码签名后，通过这个接口签名提交回来。
func (ls LoginServer) WPluginLogin(ctx context.Context, req *connect.Request[mproto.WPluginLoginRequest]) (*connect.Response[mproto.WPluginLoginResponse], error) {
	err := mmlogin.Apps.Auth.AuthorizeOnly(ctx, auth.NewAuthorizeInput(req.Msg.Address, req.Msg.Sign))
	if err != nil {
		return nil, err
	}
	res := connect.NewResponse(&mproto.WPluginLoginResponse{Token: "12324"})
	return res, nil
}
