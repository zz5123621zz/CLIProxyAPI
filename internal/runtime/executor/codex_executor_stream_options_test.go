package executor

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestApplyCodexStreamOptionsAllowlist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		body         string
		original     string
		from         sdktranslator.Format
		wantDelivery bool
	}{
		{
			name:         "allows exact value from OpenAI Responses source",
			body:         `{"model":"gpt-5","stream":true}`,
			original:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			from:         sdktranslator.FormatOpenAIResponse,
			wantDelivery: true,
		},
		{
			name:         "strips every non-allowlisted sibling",
			body:         `{"stream_options":{"include_usage":true,"other":"value"}}`,
			original:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff","include_usage":true}}`,
			from:         sdktranslator.FormatOpenAIResponse,
			wantDelivery: true,
		},
		{
			name:     "strips unsupported value",
			body:     `{"stream_options":{"reasoning_summary_delivery":"unsupported"}}`,
			original: `{"stream_options":{"reasoning_summary_delivery":"unsupported"}}`,
			from:     sdktranslator.FormatOpenAIResponse,
		},
		{
			name:     "strips option from another source format",
			body:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			original: `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			from:     sdktranslator.FormatOpenAI,
		},
		{
			name:     "strips malformed stream options",
			body:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			original: `{"stream_options":"sequential_cutoff"}`,
			from:     sdktranslator.FormatOpenAIResponse,
		},
		{
			name:     "does not trust translated or configured body without original opt in",
			body:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			original: `{"model":"gpt-5"}`,
			from:     sdktranslator.FormatOpenAIResponse,
		},
		{
			name:     "rejects malformed original JSON",
			body:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			original: `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}`,
			from:     sdktranslator.FormatOpenAIResponse,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := applyCodexStreamOptionsAllowlist([]byte(tc.body), []byte(tc.original), tc.from)
			streamOptions := gjson.GetBytes(got, "stream_options")
			if !tc.wantDelivery {
				if streamOptions.Exists() {
					t.Fatalf("stream_options survived without an allowed opt in: %s", got)
				}
				return
			}
			if !streamOptions.IsObject() {
				t.Fatalf("stream_options = %s, want object; body: %s", streamOptions.Raw, got)
			}
			if options := streamOptions.Map(); len(options) != 1 {
				t.Fatalf("stream_options has %d fields, want 1; body: %s", len(options), got)
			}
			delivery := streamOptions.Get("reasoning_summary_delivery")
			if delivery.Type != gjson.String || delivery.String() != codexReasoningSummaryDeliverySequentialCutoff {
				t.Fatalf("delivery = %s, want %q; body: %s", delivery.Raw, codexReasoningSummaryDeliverySequentialCutoff, got)
			}
		})
	}
}

func TestCodexExecutorCountTokensAlwaysStripsStreamOptions(t *testing.T) {
	t.Parallel()

	withoutStreamOptions := []byte(`{"model":"gpt-5","input":"hello"}`)
	withStreamOptions := []byte(`{"model":"gpt-5","input":"hello","stream_options":{"reasoning_summary_delivery":"sequential_cutoff","include_usage":true,"other":"this must never be counted"}}`)
	translated := sdktranslator.TranslateRequest(
		sdktranslator.FormatCodex,
		sdktranslator.FormatCodex,
		"gpt-5",
		withStreamOptions,
		false,
	)
	if !gjson.GetBytes(translated, "stream_options").Exists() {
		t.Fatal("test precondition failed: source translation removed stream_options")
	}

	executor := NewCodexExecutor(&config.Config{})
	opts := cliproxyexecutor.Options{
		SourceFormat:   sdktranslator.FormatCodex,
		ResponseFormat: sdktranslator.FormatCodex,
	}
	withoutResponse, errWithout := executor.CountTokens(context.Background(), nil, cliproxyexecutor.Request{
		Model:   "gpt-5",
		Payload: withoutStreamOptions,
	}, opts)
	if errWithout != nil {
		t.Fatalf("CountTokens() without stream_options error = %v", errWithout)
	}
	withResponse, errWith := executor.CountTokens(context.Background(), nil, cliproxyexecutor.Request{
		Model:   "gpt-5",
		Payload: withStreamOptions,
	}, opts)
	if errWith != nil {
		t.Fatalf("CountTokens() with stream_options error = %v", errWith)
	}
	if !bytes.Equal(withResponse.Payload, withoutResponse.Payload) {
		t.Fatalf(
			"stream_options changed token count: with=%s without=%s",
			withResponse.Payload,
			withoutResponse.Payload,
		)
	}
}

func TestCodexExecutorStreamOptionsAllowlistReachesHTTPBody(t *testing.T) {
	transports := []struct {
		name   string
		stream bool
	}{
		{name: "Execute", stream: false},
		{name: "ExecuteStream", stream: true},
	}
	scenarios := []struct {
		name           string
		original       string
		injectByConfig bool
		wantDelivery   bool
	}{
		{
			name:         "restores original client opt in and strips siblings",
			original:     `{"model":"gpt-5.6-sol","input":"hello","stream_options":{"reasoning_summary_delivery":"sequential_cutoff","include_usage":true}}`,
			wantDelivery: true,
		},
		{
			name:           "strips configuration injected opt in",
			original:       `{"model":"gpt-5.6-sol","input":"hello"}`,
			injectByConfig: true,
		},
	}

	for _, transport := range transports {
		transport := transport
		t.Run(transport.name, func(t *testing.T) {
			for _, scenario := range scenarios {
				scenario := scenario
				t.Run(scenario.name, func(t *testing.T) {
					upstreamBody := make(chan []byte, 1)
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						body, _ := io.ReadAll(r.Body)
						upstreamBody <- body
						w.Header().Set("Content-Type", "text/event-stream")
						_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\",\"object\":\"response\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":0,\"output_tokens\":0,\"total_tokens\":0}}}\n\n"))
					}))
					defer server.Close()

					cfg := &config.Config{}
					if scenario.injectByConfig {
						cfg.Payload = config.PayloadConfig{
							Override: []config.PayloadRule{{
								Models: []config.PayloadModelRule{{Name: "gpt-5.6-sol", Protocol: "codex"}},
								Params: map[string]any{
									"max_output_tokens":                         123,
									"stream_options.include_usage":              true,
									"stream_options.reasoning_summary_delivery": codexReasoningSummaryDeliverySequentialCutoff,
								},
							}},
						}
					}
					executor := NewCodexExecutor(cfg)
					auth := &cliproxyauth.Auth{
						Provider: "codex",
						Attributes: map[string]string{
							"api_key":   "test",
							"base_url":  server.URL,
							"plan_type": "pro",
						},
					}
					request := cliproxyexecutor.Request{
						Model:   "gpt-5.6-sol",
						Payload: []byte(`{"model":"gpt-5.6-sol","input":"hello"}`),
					}
					opts := cliproxyexecutor.Options{
						SourceFormat:    sdktranslator.FormatOpenAIResponse,
						OriginalRequest: []byte(scenario.original),
						Stream:          transport.stream,
					}

					if transport.stream {
						result, errExecute := executor.ExecuteStream(context.Background(), auth, request, opts)
						if errExecute != nil {
							t.Fatalf("ExecuteStream() error = %v", errExecute)
						}
						for chunk := range result.Chunks {
							if chunk.Err != nil {
								t.Fatalf("stream chunk error = %v", chunk.Err)
							}
						}
					} else {
						if _, errExecute := executor.Execute(context.Background(), auth, request, opts); errExecute != nil {
							t.Fatalf("Execute() error = %v", errExecute)
						}
					}

					got := <-upstreamBody
					if scenario.injectByConfig {
						if maxOutputTokens := gjson.GetBytes(got, "max_output_tokens"); maxOutputTokens.Int() != 123 {
							t.Fatalf("payload override did not run; max_output_tokens = %s, want 123; body: %s", maxOutputTokens.Raw, got)
						}
					}
					streamOptions := gjson.GetBytes(got, "stream_options")
					if !scenario.wantDelivery {
						if streamOptions.Exists() {
							t.Fatalf("upstream stream_options survived without client opt in: %s", got)
						}
						return
					}
					if options := streamOptions.Map(); len(options) != 1 {
						t.Fatalf("upstream stream_options has %d fields, want 1; body: %s", len(options), got)
					}
					if delivery := streamOptions.Get("reasoning_summary_delivery"); delivery.String() != codexReasoningSummaryDeliverySequentialCutoff {
						t.Fatalf("upstream delivery = %s, want %q; body: %s", delivery.Raw, codexReasoningSummaryDeliverySequentialCutoff, got)
					}
				})
			}
		})
	}
}
