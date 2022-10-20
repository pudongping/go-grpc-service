package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// 超时控制（上下文）
func defaultContextTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	// 通过对传入的 context 调用 ctx.Deadline 方法进行检查，若未设置截止时间的话，其将会返回 false，
	// 那么我们就会对其调用 context.WithTimeout 方法设置默认超时时间为 60 秒
	// （该超时时间设置是针对整条调用链路的，若需要另外调整，可在应用代码中再自行调整）
	if _, ok := ctx.Deadline(); !ok {
		defaultTimeout := 60 * time.Second
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	}

	return ctx, cancel
}

// UnaryContextTimeout 一元调用客户端拦截器
func UnaryContextTimeout() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := defaultContextTimeout(ctx)
		if cancel != nil {
			defer cancel()
		}

		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

// StreamContextTimeout 流式调用客户端拦截器
func StreamContextTimeout() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, cancel := defaultContextTimeout(ctx)
		if cancel != nil {
			defer cancel()
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}
