# CPA Upgrade

This context defines the language used to upgrade the CPA fork without confusing externally required compatibility with preservation of obsolete implementation details.

## Language

**Upgrade Compatibility Contract**:
The externally observable behavior and operational state that must survive a CPA upgrade: public API compatibility, production configuration and credentials, and the progressive reasoning-summary allowlist. It does not require preserving the old implementation or known defects.
_Avoid_: Development Status, Exact Version Parity

**CPA Consumer**:
Any existing application, agent, or service that relies on CPA's public interface. Ordinary API, authentication, and streaming compatibility apply to the whole consumer set, not only to La4RainGPT.
_Avoid_: La4RainGPT (when referring to every CPA client)

**Progressive Summary Compatibility**:
The La4RainGPT-specific capability to receive safe reasoning-summary sections progressively through its existing Responses transport. It is not a compatibility guarantee for every CPA Consumer or transport.
_Avoid_: Raw Thinking, WebSocket Parity

**Production Credential Set**:
The provider credentials used to serve live CPA Consumers. Exactly one CPA instance may use or persist this set at a time; parallel candidates require dedicated canary credentials.
_Avoid_: Shared Auth Directory, Copied Production Credentials

**Core Upgrade Release**:
The bounded release that upgrades CPA Core while preserving the Upgrade Compatibility Contract and leaving consumer-application and management-console versions unchanged.
_Avoid_: Platform Upgrade, Full-stack Upgrade

**Upgrade Target**:
The immutable upstream CPA revision selected for one Core Upgrade Release. For this release it is `v7.2.115` at `ffdb9c9fbc78a6235d59c9ccbdc4243ba35ecdcd`, not whichever tag is newest later.
_Avoid_: Latest Version, Moving Target

**Compatibility Gate**:
The complete body of evidence required before an upgrade candidate may become production. A missing or failed mandatory check blocks release; compilation alone is not sufficient evidence.
_Avoid_: Build Success, Best-effort Validation

**Baseline Mode**:
The initial production state of a Core Upgrade Release in which Progressive Summary Compatibility is disabled so ordinary CPA compatibility can be evaluated independently.
_Avoid_: Feature Failure, Final Operating Mode

**Capability Disable**:
The first response to a failure isolated to Progressive Summary Compatibility. It stops new opt-in requests while leaving a compatible CPA Core in service.
_Avoid_: Core Rollback

**Core Rollback**:
The restoration of the previous pinned CPA Core when ordinary API, authentication, tool, completion, consumer, or resource behavior is unsafe.
_Avoid_: Capability Disable, Retry

**Release Artifact**:
An immutable CPA image produced by the approved hosted CI from the frozen Upgrade Target and linked to its test and scan evidence.
_Avoid_: Latest Tag, Local Build

**Upstream Equivalence**:
Evidence that an official CPA release preserves the same externally observable contract as the minimal fork and passes the same Compatibility Gate. Similar-looking source code alone is not equivalence.
_Avoid_: Feature Present, Patch Looks Similar
