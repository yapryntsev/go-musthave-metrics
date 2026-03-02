package grpc

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func SubnetInterceptor(cidr string, log *zap.Logger) grpc.UnaryServerInterceptor {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Error("failed to parse cidr string: %w", zap.Error(err))
		return func(
			ctx context.Context,
			req any,
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (resp any, err error) {
			return handler(ctx, req)
		}
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "x-real-api is missing")
		}

		val := md.Get("x-real-api")
		if len(val) == 0 {
			return nil, status.Error(codes.PermissionDenied, "x-real-api is missing")
		}

		ip := net.ParseIP(val[0])
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "x-real-api contains invalid IP format")
		}

		if !subnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "untrusted subnet")
		}

		return handler(ctx, req)
	}
}
