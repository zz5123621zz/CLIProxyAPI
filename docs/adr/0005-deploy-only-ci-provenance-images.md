---
status: accepted
---

# Deploy only CI-provenance images

Production may deploy only an immutable GHCR image built, tested, and scanned by GitHub-hosted CI from the frozen Upgrade Target plus reviewed custom commits. Compose pins the full image digest; local builds and mutable tags such as `latest` are excluded so the running binary remains traceable to its source and Compatibility Gate evidence without consuming the constrained production host.
