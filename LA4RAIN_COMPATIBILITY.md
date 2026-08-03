# La4Rain progressive summary compatibility

This upgrade branch forward-ports the minimal La4Rain compatibility patch onto
upstream CLIProxyAPI `v7.2.115`, commit
`ffdb9c9fbc78a6235d59c9ccbdc4243ba35ecdcd`.

## Compatibility contract

The Codex HTTP executor continues to remove `stream_options` by default. On its
ordinary and streaming OpenAI Responses paths only, it restores exactly this
allowlisted field from the original client payload:

```json
{
  "stream_options": {
    "reasoning_summary_delivery": "sequential_cutoff"
  }
}
```

No sibling key, alternative value, malformed object, value injected by payload
configuration, or field from another source format is forwarded. The
token-count path continues to remove all stream options. This fork deliberately
does not change the Codex WebSocket executors, so Progressive Summary
Compatibility is not guaranteed on WebSocket routes.

The option asks the upstream Codex service to deliver completed safe reasoning
summary sections progressively. It does not expose raw chain of thought and
does not alter model selection, reasoning effort, tools, image parameters,
authentication, or quota handling.

## Frozen baseline CI conformance

The unmodified `v7.2.115` target has a nondeterministic Home plugin-sync
cancellation path and fails Go 1.26 vet in pre-existing plugin stream-lifecycle
and request-logging code. This integration branch keeps the complete test and
vet gates and includes semantics-preserving conformance fixes:

- failed stream setup explicitly cancels its context, while a successful bridge
  retains the existing cancellation ownership;
- cancellation of a dedicated Home plugin-sync request closes its connection so
  a later Redis read deadline cannot mask the cancellation;
- the internal `FileBodySource.WriteTo` helper is named `WriteBodyTo` so it is
  not mistaken for the standard `io.WriterTo` method with a different
  signature.

These fixes do not change the Progressive Summary Compatibility scope or any
public CPA protocol.

## Release gate

The dedicated GitHub-hosted workflow runs the complete Go test and vet suites,
refreshes the model catalog, builds the server and `linux/amd64` image, and
scans the image for HIGH/CRITICAL vulnerabilities before the run can pass. It
publishes to `ghcr.io/zz5123621zz/cliproxyapi-la4rain`; production must pin the
digest from a successful run, never a mutable tag or local build.

Only one CPA instance may use or persist the production credential set. A
parallel canary must use separate credentials, and the old instance must stop
before the new instance receives production OAuth credentials.

## Activation and rollback

The Core Upgrade Release keeps La4RainGPT and CPA Manager Plus on their current
versions. It starts with progressive summaries set to `off`; after ordinary
API, authentication, streaming, tool, completion, consumer, and resource checks
pass, the capability may switch to `auto`.

A failure isolated to progressive summaries disables that capability. Any
failure in ordinary requests, authentication refresh, tools, completion events,
another CPA consumer, credential integrity, or resource stability rolls the
whole CPA Core back to the previously pinned image.

The fork can be retired only after an official CPA release passes the same
compatibility gate and proves equivalent externally observable behavior.
