package grpcserver

import (
	"context"

	"link-shortener/internal/auth"
	"link-shortener/internal/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthFunc func(ctx context.Context) (context.Context, error)

func UnaryServerInterceptor(authFunc AuthFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		newCtx, err := authFunc(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

func Authenticate(ctx context.Context) (context.Context, error) {
	var userID string

	vals := metadata.ValueFromIncomingContext(ctx, "authorization")
	if len(vals) > 0 {
		parsedUserID, valid := auth.ParseCookieValue(vals[0])
		if valid {
			userID = parsedUserID
		}
	}

	if userID == "" {
		newUserID, err := auth.GenerateUserID()
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to generate user id")
		}
		userID = newUserID
	}

	newCtx := context.WithValue(ctx, middleware.UserIDKey, userID)
	return newCtx, nil
}
