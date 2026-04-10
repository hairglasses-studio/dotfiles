package clients

import (
	"context"

	mcpobs "github.com/hairglasses-studio/mcpkit/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const ollamaTracerName = "github.com/hairglasses-studio/hg-mcp/internal/clients/ollama"

var ollamaServerAddressAttr = attribute.Key("server.address")

func startOllamaLLMSpan(ctx context.Context, operation, model, baseURL string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		mcpobs.AttrGenAISystem.String("ollama"),
		mcpobs.AttrGenAIOperationName.String(operation),
	}
	if model != "" {
		attrs = append(attrs, mcpobs.AttrGenAIRequestModel.String(model))
	}
	if baseURL != "" {
		attrs = append(attrs, ollamaServerAddressAttr.String(baseURL))
	}

	return otel.Tracer(ollamaTracerName).Start(
		ctx,
		"ollama."+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
}

func finishOllamaLLMSpan(span trace.Span, model string, inputTokens, outputTokens int, err error) {
	if span == nil {
		return
	}

	if model != "" {
		span.SetAttributes(mcpobs.AttrGenAIRequestModel.String(model))
	}
	if inputTokens > 0 {
		span.SetAttributes(mcpobs.AttrGenAIUsageInput.Int(inputTokens))
	}
	if outputTokens > 0 {
		span.SetAttributes(mcpobs.AttrGenAIUsageOutput.Int(outputTokens))
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}
