// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_recovery

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFromPanic(ctx, r, o.recoveryHandlerFunc)
				if ctx.Err() != nil {
					err = status.FromContextError(ctx.Err()).Err()
				}
			}
		}()
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor for panic recovery.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateOptions(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFromPanic(stream.Context(), r, o.recoveryHandlerFunc)
				if stream.Context().Err() != nil {
					err = status.FromContextError(stream.Context().Err()).Err()
				}
			}
		}()
		return handler(srv, stream)
	}
}

func recoverFromPanic(ctx context.Context, r interface{}, recoveryHandlerFunc RecoveryHandlerFuncContext) error {
	if recoveryHandlerFunc != nil {
		return recoveryHandlerFunc(ctx, r)
	}
	return status.Errorf(codes.Internal, "%v", r)
}