package executor

import (
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const codexReasoningSummaryDeliverySequentialCutoff = "sequential_cutoff"

// applyCodexStreamOptionsAllowlist removes every stream option, then restores
// the one provider-supported option when an OpenAI Responses client explicitly
// supplied it in the original request. Translated or configuration-injected
// values are never trusted.
func applyCodexStreamOptionsAllowlist(body, originalPayloadSource []byte, from sdktranslator.Format) []byte {
	sanitized, _ := sjson.DeleteBytes(body, "stream_options")
	if from != sdktranslator.FormatOpenAIResponse || !gjson.ValidBytes(originalPayloadSource) {
		return sanitized
	}

	streamOptions := gjson.GetBytes(originalPayloadSource, "stream_options")
	if !streamOptions.IsObject() {
		return sanitized
	}
	delivery := streamOptions.Get("reasoning_summary_delivery")
	if delivery.Type != gjson.String || delivery.String() != codexReasoningSummaryDeliverySequentialCutoff {
		return sanitized
	}

	updated, err := sjson.SetBytes(
		sanitized,
		"stream_options.reasoning_summary_delivery",
		codexReasoningSummaryDeliverySequentialCutoff,
	)
	if err != nil {
		return sanitized
	}
	return updated
}
