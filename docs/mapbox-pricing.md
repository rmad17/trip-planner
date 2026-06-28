# Mapbox Pricing & Billing Reference

> **Last verified:** 2026-06-28 against <https://www.mapbox.com/pricing> and the
> Mapbox API docs. Pricing tiers change — re-verify against the live pricing
> page and your account dashboard (Statistics → Usage) before relying on exact
> numbers. This doc is the project's single source of truth for *why* we chose
> the Mapbox APIs we use.

## TL;DR — decisions for this project

- **Geocoding:** Use **Geocoding API v6 forward** (`/search/geocode/v6/forward`)
  with `permanent` left at its default (`false`). This bills as **Temporary
  Geocoding** — 100,000 requests/month free, no minimum charge.
- **Never** set `permanent=true` and never call the `mapbox.places-permanent`
  (v5) endpoint. **Permanent Geocoding has no free tier and a $5/1k minimum**,
  billed from the very first request. This is what caused the account `rmad17`
  overage notice.
- **Directions:** Mapbox Directions API (`/directions/v5`) is fine — 100k/mo
  free, per request. It is currently a *fallback* (default routes provider is
  Google via `ROUTES_PROVIDER`).
- **Avoid** the standalone Search Box sessions model for our use case: its free
  tier is only ~500 sessions/month and it requires correct `session_token`
  correlation to bill efficiently.

## The Permanent Geocoding trap (why we migrated)

| | Temporary Geocoding | Permanent Geocoding |
|---|---|---|
| v6 form | `permanent=false` (default) | `permanent=true` |
| v5 endpoint | `mapbox.places` | `mapbox.places-permanent` |
| Free tier | 100,000 req/mo | **None** |
| Price | $0.75/1k and down | **$5.00/1k, $5 minimum** |
| May store results? | No | Yes |

Permanent geocoding bills **from request #1** with a **$5/month minimum
commitment**. Even a handful of requests put the account on this line item.

## Full pricing tables

Mapbox lists "introductory" pricing (new/eligible accounts) and higher
"standard" pricing. Confirm which applies to your account in the dashboard.

### Search products

| Product | Free tier | Unit | First paid tier |
|---|---|---|---|
| **Temporary Geocoding API** | 100,000 / mo | request | $0.75/1k (100k–500k), $0.60/1k (500k–1M), $0.45/1k (1M+) |
| **Permanent Geocoding API** | **0** | request | **$5.00/1k (1–500k, $5 min)**, $4.00/1k (500k+) |
| **Search Box API — Sessions** (`/suggest`+`/retrieve`) | 500 / mo | session (≤50 suggest + 1 retrieve) | intro $3.00/1k; standard $11.50/1k |
| **Search Box API — Requests** (`/category`, `/reverse`) | 50,000 / mo | request | intro $1.00/1k; standard $1.70/1k |
| **Address Autofill** | 1,000 / mo | session | $12.50/1k (1k–25k) |

### Navigation products

| Product | Free tier | Unit | First paid tier |
|---|---|---|---|
| **Directions API** | 100,000 / mo | request | $2.00/1k (100k–500k), $1.60/1k (500k–1M), $1.20/1k (1M+) |
| **Matrix API** | 100,000 / mo | matrix element | $2.00/1k (100k–500k) |
| **Isochrone API** | 100,000 / mo | request | $2.00/1k (100k–500k) |

### Maps & tiles

| Product | Free tier | Unit | First paid tier |
|---|---|---|---|
| **Mapbox GL JS (Map Loads)** | 50,000 / mo | map load | $5.00/1k (50k–100k) … $2.50/1k (1M–5M) |
| **Static Images API** | 50,000 / mo | request | $1.00/1k (50k–500k) |
| **Vector Tiles API** | 200,000 / mo | tile request | $0.25/1k (200k–2M) |
| **Raster Tiles API** | 750,000 / mo | tile request | $0.25/1k (750k–2M) |

## What this project actually uses

| Code | Mapbox API | Bills as | Free tier | Notes |
|---|---|---|---|---|
| `places/api.go`, `places/mapbox.go` | **→ migrating to Geocoding v6 forward** (was Search Box `/suggest`+`/retrieve`) | **Temporary Geocoding** | 100k/mo | `permanent` must stay unset/false |
| `routes/mapbox.go` | Directions API v5 (`mapbox/driving|walking|cycling`) | Directions | 100k/mo | Fallback only; default routes provider is Google |

Not used (confirmed absent): Map Loads / GL JS, Static Images, Matrix,
Isochrone, Tilequery, Reverse/Category search, `permanent=true` anywhere.

## Guardrails

- Keep `permanent` **unset** on every geocoding call. If you ever need to store
  results, treat that as a deliberate, costed decision — not a default.
- Use the standard (temporary) `mapbox.places` / v6 default endpoints only.
- Set a **spending limit + usage alert** in the Mapbox dashboard.
- Debounce autocomplete on the client (~300 ms) and cache results by
  `mapbox_id` to stay well inside the 100k/mo free tier.
- Periodically check **Dashboard → Statistics**, filtered by token + endpoint,
  to confirm nothing is landing on the Permanent Geocoding line.
