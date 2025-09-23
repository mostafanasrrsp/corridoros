# Liminal Research Ethics — CorridorLabs Only

Purpose
- Establishes consent-first, women-led guardrails for any human synchrony work.
- Applies only to CorridorLabs modules (offline/simulation). CorridorOS production services do not capture human data.

Principles
- Consent-first: No data ingestion without explicit, informed, revocable consent per participant.
- Women-led: Any field deployment requires women-led governance and review; defaults OFF.
- No sacred spaces: No sensors in mourning rooms or sacred spaces unless the community invites and leads.
- Minimal data: Collect only anonymized, aggregate signals needed for specific study; default to offline analysis.
- Transparency: Clear flags labeling “Hypothesis” vs “Validated” metrics; provenance and versioning for analytics.
- Auditability: Every dataset includes a Consent Manifest and an Ethics Attestation embedded alongside metrics.

Operational Model
- CorridorLabs-only services provide:
  - Offline/simulation analytics for group synchrony (breath/HRV) with anonymized inputs.
  - Consent Manifest schema, enforced at API boundaries (requests rejected without it).
  - Women-led governance toggle that must be true to enable analysis functions.
- Production CorridorOS (corrd, memqosd, fabmand, etc.) contain no human-signal capture logic.

Consent Manifest (required)
- Required fields:
  - `study_id` (string), `version` (string)
  - `community_governance` with `{ women_led: true, contact: string }`
  - `participants[]` with `{ pseudonym: string, consent: true, scope: ["offline_analysis"], retention_days: int }`
  - `data_minimization`: true | false (must be true)
  - `capture_mode`: "offline" | "realtime" (CorridorLabs supports offline only)
  - `timestamp` (RFC3339)

Ethics Attestation
- Each analytics run produces an attestation companion record:
  - `attestation_id`, `study_id`, `manifest_hash`, `tooling_version`, `flags: ["simulation"|"offline"]`.
  - Signed with PQC (Dilithium) when configured.

Non‑Permitted (Gen 1)
- Realtime or covert collection in sensitive spaces.
- Face/video biometrics, voice prints, or direct identifiers.
- Data sharing outside consent scope or without cryptographic attestation.

Review & Redress
- Provide community contact, revocation path, and deletion SLAs equal to or less than `retention_days`.

Labeling
- CorridorLabs UIs and APIs must label: “Simulation”, “Offline”, and “Women‑led governance required” when applicable.

