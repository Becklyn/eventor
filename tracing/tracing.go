package tracing

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func FromFiberContext(ctx *fiber.Ctx) context.Context {
	traceparent := ctx.Get("traceparent")
	if len(traceparent) == 0 {
		return ctx.Context()
	}
	traceparentParts := strings.Split(traceparent, "-")
	traceId, err := trace.TraceIDFromHex(traceparentParts[1])
	if err != nil {
		return ctx.Context()
	}
	spanId, err := trace.SpanIDFromHex(traceparentParts[2])
	if err != nil {
		return ctx.Context()
	}

	traceFlagsInt, err := strconv.Atoi(traceparentParts[3])
	if err != nil {
		return ctx.Context()
	}
	traceFlags := trace.TraceFlags(traceFlagsInt)

	tracestate := ctx.Get("tracestate")
	paresedTracestate, err := trace.ParseTraceState(tracestate)
	if err != nil {
		return ctx.Context()
	}

	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: traceFlags,
		TraceState: paresedTracestate,
		Remote:     true,
	})
	return trace.ContextWithRemoteSpanContext(ctx.Context(), spanCtx)
}

func ToHeaders(spanCtx trace.SpanContext) (string, string) {
	traceparent := fmt.Sprintf("00-%s-%s-%s",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		spanCtx.TraceFlags().String(),
	)

	tracestate := spanCtx.TraceState().String()

	return traceparent, tracestate
}
