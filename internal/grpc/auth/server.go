package auth

import (
	"context"
	ssov1 "github.com/rinefica/voice_null_protos/gen/go/sso"
	"github.com/rinefica/voice_null_sso/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth auth.AuthService
}

func Register(gRPC *grpc.Server, auth auth.AuthService) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}
	token, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}
	userID, err := s.auth.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if (req.GetEmail() == "") || (req.GetPassword() == "") {
		return status.Error(codes.InvalidArgument, "email or password is required")
	}
	if req.AppId < 1 {
		return status.Error(codes.InvalidArgument, "app id should be greater than zero")
	}
	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if (req.GetEmail() == "") || (req.GetPassword() == "") {
		return status.Error(codes.InvalidArgument, "email or password is required")
	}
	return nil
}
