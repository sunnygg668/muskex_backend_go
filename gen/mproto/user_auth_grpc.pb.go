// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: mproto/user_auth.proto

package mproto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	UserAuthService_VerifyCode_FullMethodName         = "/mproto.UserAuthService/VerifyCode"
	UserAuthService_SignUp_FullMethodName             = "/mproto.UserAuthService/SignUp"
	UserAuthService_SignIn_FullMethodName             = "/mproto.UserAuthService/SignIn"
	UserAuthService_ForgetPwd_FullMethodName          = "/mproto.UserAuthService/ForgetPwd"
	UserAuthService_UpdateFundPassword_FullMethodName = "/mproto.UserAuthService/UpdateFundPassword"
)

// UserAuthServiceClient is the client API for UserAuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserAuthServiceClient interface {
	// 验证码 verify code
	VerifyCode(ctx context.Context, in *VerifyCodeRequest, opts ...grpc.CallOption) (*VerifyCodeResponse, error)
	// 用户注册 user regist
	SignUp(ctx context.Context, in *SignUpRequest, opts ...grpc.CallOption) (*SignUpResponse, error)
	// 用户登陆
	SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error)
	// 短信验证取回登陆密码
	ForgetPwd(ctx context.Context, in *ForgetPwdRequest, opts ...grpc.CallOption) (*MsgResponse, error)
	// 短信验证修改资金密码
	UpdateFundPassword(ctx context.Context, in *ForgetPwdRequest, opts ...grpc.CallOption) (*MsgResponse, error)
}

type userAuthServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserAuthServiceClient(cc grpc.ClientConnInterface) UserAuthServiceClient {
	return &userAuthServiceClient{cc}
}

func (c *userAuthServiceClient) VerifyCode(ctx context.Context, in *VerifyCodeRequest, opts ...grpc.CallOption) (*VerifyCodeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VerifyCodeResponse)
	err := c.cc.Invoke(ctx, UserAuthService_VerifyCode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userAuthServiceClient) SignUp(ctx context.Context, in *SignUpRequest, opts ...grpc.CallOption) (*SignUpResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SignUpResponse)
	err := c.cc.Invoke(ctx, UserAuthService_SignUp_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userAuthServiceClient) SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SignInResponse)
	err := c.cc.Invoke(ctx, UserAuthService_SignIn_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userAuthServiceClient) ForgetPwd(ctx context.Context, in *ForgetPwdRequest, opts ...grpc.CallOption) (*MsgResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgResponse)
	err := c.cc.Invoke(ctx, UserAuthService_ForgetPwd_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userAuthServiceClient) UpdateFundPassword(ctx context.Context, in *ForgetPwdRequest, opts ...grpc.CallOption) (*MsgResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgResponse)
	err := c.cc.Invoke(ctx, UserAuthService_UpdateFundPassword_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserAuthServiceServer is the server API for UserAuthService service.
// All implementations must embed UnimplementedUserAuthServiceServer
// for forward compatibility.
type UserAuthServiceServer interface {
	// 验证码 verify code
	VerifyCode(context.Context, *VerifyCodeRequest) (*VerifyCodeResponse, error)
	// 用户注册 user regist
	SignUp(context.Context, *SignUpRequest) (*SignUpResponse, error)
	// 用户登陆
	SignIn(context.Context, *SignInRequest) (*SignInResponse, error)
	// 短信验证取回登陆密码
	ForgetPwd(context.Context, *ForgetPwdRequest) (*MsgResponse, error)
	// 短信验证修改资金密码
	UpdateFundPassword(context.Context, *ForgetPwdRequest) (*MsgResponse, error)
	mustEmbedUnimplementedUserAuthServiceServer()
}

// UnimplementedUserAuthServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedUserAuthServiceServer struct{}

func (UnimplementedUserAuthServiceServer) VerifyCode(context.Context, *VerifyCodeRequest) (*VerifyCodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyCode not implemented")
}
func (UnimplementedUserAuthServiceServer) SignUp(context.Context, *SignUpRequest) (*SignUpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignUp not implemented")
}
func (UnimplementedUserAuthServiceServer) SignIn(context.Context, *SignInRequest) (*SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}
func (UnimplementedUserAuthServiceServer) ForgetPwd(context.Context, *ForgetPwdRequest) (*MsgResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ForgetPwd not implemented")
}
func (UnimplementedUserAuthServiceServer) UpdateFundPassword(context.Context, *ForgetPwdRequest) (*MsgResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateFundPassword not implemented")
}
func (UnimplementedUserAuthServiceServer) mustEmbedUnimplementedUserAuthServiceServer() {}
func (UnimplementedUserAuthServiceServer) testEmbeddedByValue()                         {}

// UnsafeUserAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserAuthServiceServer will
// result in compilation errors.
type UnsafeUserAuthServiceServer interface {
	mustEmbedUnimplementedUserAuthServiceServer()
}

func RegisterUserAuthServiceServer(s grpc.ServiceRegistrar, srv UserAuthServiceServer) {
	// If the following call pancis, it indicates UnimplementedUserAuthServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&UserAuthService_ServiceDesc, srv)
}

func _UserAuthService_VerifyCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyCodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserAuthServiceServer).VerifyCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserAuthService_VerifyCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserAuthServiceServer).VerifyCode(ctx, req.(*VerifyCodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserAuthService_SignUp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignUpRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserAuthServiceServer).SignUp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserAuthService_SignUp_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserAuthServiceServer).SignUp(ctx, req.(*SignUpRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserAuthService_SignIn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignInRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserAuthServiceServer).SignIn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserAuthService_SignIn_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserAuthServiceServer).SignIn(ctx, req.(*SignInRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserAuthService_ForgetPwd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ForgetPwdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserAuthServiceServer).ForgetPwd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserAuthService_ForgetPwd_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserAuthServiceServer).ForgetPwd(ctx, req.(*ForgetPwdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserAuthService_UpdateFundPassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ForgetPwdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserAuthServiceServer).UpdateFundPassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserAuthService_UpdateFundPassword_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserAuthServiceServer).UpdateFundPassword(ctx, req.(*ForgetPwdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserAuthService_ServiceDesc is the grpc.ServiceDesc for UserAuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserAuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mproto.UserAuthService",
	HandlerType: (*UserAuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "VerifyCode",
			Handler:    _UserAuthService_VerifyCode_Handler,
		},
		{
			MethodName: "SignUp",
			Handler:    _UserAuthService_SignUp_Handler,
		},
		{
			MethodName: "SignIn",
			Handler:    _UserAuthService_SignIn_Handler,
		},
		{
			MethodName: "ForgetPwd",
			Handler:    _UserAuthService_ForgetPwd_Handler,
		},
		{
			MethodName: "UpdateFundPassword",
			Handler:    _UserAuthService_UpdateFundPassword_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mproto/user_auth.proto",
}

const (
	WpAuthService_WPChallenge_FullMethodName  = "/mproto.WpAuthService/WPChallenge"
	WpAuthService_WPluginLogin_FullMethodName = "/mproto.WpAuthService/WPluginLogin"
)

// WpAuthServiceClient is the client API for WpAuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// okx tron钱包插件登陆
type WpAuthServiceClient interface {
	// okx tron钱包插件登陆：app获取登陆授权码，后续app把授权码签名后，会把签名提交回来。
	WPChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*ChallengeResponse, error)
	// okx tron钱包插件登陆(WP,wallet plugin): app把授权码签名后，通过这个接口签名提交回来,获得token。
	WPluginLogin(ctx context.Context, in *WPluginLoginRequest, opts ...grpc.CallOption) (*WPluginLoginResponse, error)
}

type wpAuthServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWpAuthServiceClient(cc grpc.ClientConnInterface) WpAuthServiceClient {
	return &wpAuthServiceClient{cc}
}

func (c *wpAuthServiceClient) WPChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*ChallengeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChallengeResponse)
	err := c.cc.Invoke(ctx, WpAuthService_WPChallenge_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wpAuthServiceClient) WPluginLogin(ctx context.Context, in *WPluginLoginRequest, opts ...grpc.CallOption) (*WPluginLoginResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WPluginLoginResponse)
	err := c.cc.Invoke(ctx, WpAuthService_WPluginLogin_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WpAuthServiceServer is the server API for WpAuthService service.
// All implementations must embed UnimplementedWpAuthServiceServer
// for forward compatibility.
//
// okx tron钱包插件登陆
type WpAuthServiceServer interface {
	// okx tron钱包插件登陆：app获取登陆授权码，后续app把授权码签名后，会把签名提交回来。
	WPChallenge(context.Context, *ChallengeRequest) (*ChallengeResponse, error)
	// okx tron钱包插件登陆(WP,wallet plugin): app把授权码签名后，通过这个接口签名提交回来,获得token。
	WPluginLogin(context.Context, *WPluginLoginRequest) (*WPluginLoginResponse, error)
	mustEmbedUnimplementedWpAuthServiceServer()
}

// UnimplementedWpAuthServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWpAuthServiceServer struct{}

func (UnimplementedWpAuthServiceServer) WPChallenge(context.Context, *ChallengeRequest) (*ChallengeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WPChallenge not implemented")
}
func (UnimplementedWpAuthServiceServer) WPluginLogin(context.Context, *WPluginLoginRequest) (*WPluginLoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WPluginLogin not implemented")
}
func (UnimplementedWpAuthServiceServer) mustEmbedUnimplementedWpAuthServiceServer() {}
func (UnimplementedWpAuthServiceServer) testEmbeddedByValue()                       {}

// UnsafeWpAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WpAuthServiceServer will
// result in compilation errors.
type UnsafeWpAuthServiceServer interface {
	mustEmbedUnimplementedWpAuthServiceServer()
}

func RegisterWpAuthServiceServer(s grpc.ServiceRegistrar, srv WpAuthServiceServer) {
	// If the following call pancis, it indicates UnimplementedWpAuthServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&WpAuthService_ServiceDesc, srv)
}

func _WpAuthService_WPChallenge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChallengeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WpAuthServiceServer).WPChallenge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WpAuthService_WPChallenge_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WpAuthServiceServer).WPChallenge(ctx, req.(*ChallengeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WpAuthService_WPluginLogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WPluginLoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WpAuthServiceServer).WPluginLogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WpAuthService_WPluginLogin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WpAuthServiceServer).WPluginLogin(ctx, req.(*WPluginLoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WpAuthService_ServiceDesc is the grpc.ServiceDesc for WpAuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WpAuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mproto.WpAuthService",
	HandlerType: (*WpAuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WPChallenge",
			Handler:    _WpAuthService_WPChallenge_Handler,
		},
		{
			MethodName: "WPluginLogin",
			Handler:    _WpAuthService_WPluginLogin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mproto/user_auth.proto",
}
