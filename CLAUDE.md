# CLAUDE.md

Guidance for Claude Code when working in the trip-planner repo.

## Third-party API billing — raise a flag

This project integrates many paid/limited external APIs. **Cost control is a
first-class concern.** A full registry of every provider, its free tier, and
flags lives in [`docs/api-billing.md`](docs/api-billing.md) (Mapbox detail in
[`docs/mapbox-pricing.md`](docs/mapbox-pricing.md)). Keep both up to date.

### When to raise a flag (do this proactively, unprompted)

Whenever you add, modify, or review code that calls an external API — or change
a default provider / env var that selects one — **stop and check its billing**,
and surface a clear ⚠️/🔴 flag to the user if any of these are true:

- The API has **no free tier** (pays from the first call) — e.g. Anthropic,
  OpenAI, Twilio, Mapbox **Permanent** Geocoding, DigitalOcean Spaces.
- The API has a **minimum monthly commitment** — e.g. Mapbox Permanent
  Geocoding ($5/mo min).
- The API has a **low free tier** that modest traffic could exhaust — e.g.
  **Google Maps Platform** Pro SKUs (Routes, Place Details) at ~5,000 calls/mo,
  Amadeus (~2,000/mo), SendGrid trial.
- A call is placed **in a loop / per-item** (per hop, per result, per keystroke)
  without caching or debouncing.
- A change would **switch a default** from a free/cheap provider to a paid one.

**Pay special attention to Mapbox and Google Maps** — both have changed their
billing models recently and have surprisingly low or zero free tiers on some
SKUs (Google dropped the $200 universal credit on 2025-03-01; Mapbox Permanent
Geocoding has no free tier).

### How to flag

State plainly: the provider, the billing unit, the free tier, and the risk —
e.g. "⚠️ This uses Google Routes (Pro SKU, ~5,000 free calls/mo) as the default
and is called once per hop with no caching." Recommend a cheaper path
(free-tier provider, caching, debounce) and update `docs/api-billing.md`.

## Hard rules

- **Never** use Mapbox **Permanent** geocoding: do not set `permanent=true` and
  do not call `mapbox.places-permanent`. Use Geocoding **v6** forward with the
  default `permanent=false` (bills as Temporary Geocoding, 100k/mo free).
- Keep free/cheap defaults: `DEFAULT_LLM_PROVIDER=gemini`,
  `STORAGE_PROVIDER=local`, and prefer `ROUTES_PROVIDER=mapbox` or route caching
  over Google Routes for high-volume paths.
- Any time you touch an API integration, update the billing registry.
