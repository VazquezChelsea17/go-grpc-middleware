package grpc_recovery

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptor_PanicWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	interceptor := UnaryServerInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("simulated panic after cancel")
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Code() != codes.Canceled {
		t.Errorf("expected code %v, got %v", codes.Canceled, st.Code())
	}
}

func TestUnaryServerInterceptor_PanicWithContextDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond) // Wait for deadline to expire

	interceptor := UnaryServerInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("simulated panic after deadline")
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Code() != codes.DeadlineExceeded {
		t.Errorf("expected code %v, got %v", codes.DeadlineExceeded, st.Code())
	}
}

func TestUnaryServerInterceptor_PanicWithActiveContext(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("simulated panic")
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Code() != codes.Internal {
		t.Errorf("expected code %v, got %v", codes.Internal, st.Code())
	}
}

func TestStreamServerInterceptor_PanicWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	interceptor := StreamServerInterceptor()
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		panic("simulated panic after cancel")
	}

	stream := &mockServerStream{ctx: ctx}
	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Code() != codes.Canceled {
		t.Errorf("expected code %v, got %v", codes.Canceled, st.Code())
	}
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}