---
status: accepted
---

# Use two-level rollback

A failure isolated to Progressive Summary Compatibility triggers Capability Disable while the upgraded core remains in service. Any failure in ordinary requests, authentication refresh, tool execution, terminal event delivery, another CPA Consumer, credential integrity, or resource stability triggers Core Rollback to the previous pinned image within five minutes; this separation avoids discarding a sound core upgrade for an optional capability failure without tolerating system-wide regressions.
