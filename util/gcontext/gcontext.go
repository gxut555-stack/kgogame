package gcontext

import (
	"context"
	"github.com/lithammer/shortuuid/v4"
)

const TraceId = "TraceId"

func WithTraceId(ctx context.Context) context.Context {
	if ctx == nil {
		return NewContext()
	}
	if ctx.Value(TraceId) == nil {
		ctx = context.WithValue(ctx, TraceId, shortuuid.New())
	}
	return ctx
}

func NewContext() context.Context {
	return context.WithValue(context.Background(), TraceId, shortuuid.New())
}
