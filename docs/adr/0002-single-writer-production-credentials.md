---
status: accepted
---

# Keep a single writer for production credentials

Only one CPA instance may use and persist the Production Credential Set during the upgrade. A parallel candidate must use dedicated canary credentials; final cutover stops the old instance before the new instance receives the production credentials, because separate auth directories do not prevent two instances from racing the same upstream OAuth refresh token.
