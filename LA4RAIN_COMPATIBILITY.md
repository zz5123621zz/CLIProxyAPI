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
