---
status: accepted
---

# Decouple the CPA Core upgrade

The Core Upgrade Release changes only CPA Core, the Progressive Summary Compatibility implementation, and configuration required for compatibility. La4RainGPT and CPA Manager Plus remain on their current versions; any management-console incompatibility will be handled as a separate change so that the CPA upgrade has a clear failure domain and independent rollback.
