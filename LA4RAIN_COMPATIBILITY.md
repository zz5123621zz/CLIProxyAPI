# La4Rain progressive summary compatibility

This fork is based on upstream CLIProxyAPI `v7.2.96`, commit
`285322cd97add6b21f60c267debec44fbec74060`.

## Compatibility contract

The Codex executor continues to remove `stream_options` by default. For an
OpenAI Responses source request only, it restores exactly this allowlisted
field from the original client payload:

```json
{
  "stream_options": {
    "reasoning_summary_delivery": "sequential_cutoff"
  }
}
```

No sibling key, alternative value, value injected by payload configuration, or
field from another source format is forwarded. The token-count path continues
to remove all stream options.

This option asks the upstream Codex service to deliver completed safe reasoning
summary sections progressively. It does not expose raw chain of thought and
does not alter model selection, reasoning effort, tools, image parameters, or
quota handling.

## Release and rollback

GitHub Actions tests the executor, builds the server, publishes the
`linux/amd64` image to
`ghcr.io/zz5123621zz/cliproxyapi-la4rain`, and scans it for reachable
HIGH/CRITICAL image vulnerabilities. Production must pin the resulting digest,
not a mutable tag.

If upstream CLIProxyAPI implements the same contract, La4Rain can return to an
official image without changing its response event schema. To roll back this
fork, restore the previously recorded CPA binary or image and switch
La4RainGPT's administrator setting to `off`.
