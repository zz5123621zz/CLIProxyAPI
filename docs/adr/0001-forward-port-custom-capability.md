---
status: accepted
---

# Forward-port the custom capability onto v7.2.115

The CPA upgrade will start from official `v7.2.115` at `ffdb9c9fbc78a6235d59c9ccbdc4243ba35ecdcd` on a separate integration branch and reimplement Progressive Summary Compatibility against the new executor layout. That Upgrade Target remains frozen even if a newer tag appears during integration. The production `v7.2.96` branch and pinned image remain immutable rollback baselines; the team will not merge or rebase the production branch or directly cherry-pick `99f2204`, because the old monolithic executor conflicts with the split target structure.
