package grpcserver

import (
	"context"

	pb "link-shortener/proto"

	"link-shortener/internal/middleware"
	"link-shortener/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer

	service service.ShortenerService
	baseURL string
}

func NewShortenerServer(svc service.ShortenerService, baseURL string) *ShortenerServer {
	return &ShortenerServer{
		service: svc,
		baseURL: baseURL,
	}
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID := middleware.GetUserID(ctx)

	shortURL, err := s.service.ShortenURL(ctx, in.GetUrl(), userID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid url: %v", err)
	}

	var response pb.URLShortenResponse
	response.SetResult(s.baseURL + "/" + shortURL)
	return &response, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	originalURL, err := s.service.GetOriginalURL(ctx, in.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "url not found")
	}

	var response pb.URLExpandResponse
	response.SetResult(originalURL)
	return &response, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, in *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID := middleware.GetUserID(ctx)

	urls, err := s.service.GetURLsByUserID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get urls: %v", err)
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "no urls found")
	}

	urlData := make([]*pb.URLData, len(urls))
	for i, u := range urls {
		urlData[i] = pb.URLData_builder{
			ShortUrl:    proto.String(s.baseURL + "/" + u.ShortURL),
			OriginalUrl: proto.String(u.OriginalURL),
		}.Build()
	}

	return pb.UserURLsResponse_builder{Url: urlData}.Build(), nil
}
