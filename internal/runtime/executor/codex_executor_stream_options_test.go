package executor

import (
	"testing"

	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestApplyCodexStreamOptionsAllowlist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		body             string
		original         string
		from             sdktranslator.Format
		wantDelivery     bool
		wantIncludeUsage bool
	}{
		{
			name:         "allows exact value from OpenAI Responses source",
			body:         `{"model":"gpt-5","stream":true}`,
			original:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			from:         sdktranslator.FormatOpenAIResponse,
			wantDelivery: true,
		},
		{
			name:             "strips every non-allowlisted sibling",
			body:             `{"stream_options":{"include_usage":true,"other":"value"}}`,
			original:         `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff","include_usage":true}}`,
			from:             sdktranslator.FormatOpenAIResponse,
			wantDelivery:     true,
			wantIncludeUsage: false,
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
			name:     "does not trust translated body without original opt in",
			body:     `{"stream_options":{"reasoning_summary_delivery":"sequential_cutoff"}}`,
			original: `{"model":"gpt-5"}`,
			from:     sdktranslator.FormatOpenAIResponse,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := applyCodexStreamOptionsAllowlist([]byte(tc.body), []byte(tc.original), tc.from)
			delivery := gjson.GetBytes(got, "stream_options.reasoning_summary_delivery")
			if delivery.Exists() != tc.wantDelivery {
				t.Fatalf("delivery presence = %v, want %v; body: %s", delivery.Exists(), tc.wantDelivery, got)
			}
			if tc.wantDelivery && delivery.String() != codexReasoningSummaryDeliverySequentialCutoff {
				t.Fatalf("delivery = %q, want %q", delivery.String(), codexReasoningSummaryDeliverySequentialCutoff)
			}
			includeUsage := gjson.GetBytes(got, "stream_options.include_usage")
			if includeUsage.Exists() != tc.wantIncludeUsage {
				t.Fatalf("include_usage presence = %v, want %v; body: %s", includeUsage.Exists(), tc.wantIncludeUsage, got)
			}
			if other := gjson.GetBytes(got, "stream_options.other"); other.Exists() {
				t.Fatalf("non-allowlisted stream option survived: %s", got)
			}
		})
	}
}
