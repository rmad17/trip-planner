# Third-Party API Billing Registry

> **Last verified:** 2026-06-28. Pricing/free tiers change often — re-verify
> against each provider's pricing page and your account dashboard before relying
> on exact numbers. This is the master registry of **every external paid/limited
> API the trip-planner consumes**. Mapbox detail lives in
> [`mapbox-pricing.md`](./mapbox-pricing.md).

## Flag legend

- ✅ **Comfortable free tier** — unlikely to bill at our scale.
- ⚠️ **Low / limited free tier** — watch usage; can bill with modest traffic.
- 🔴 **No free tier / pays from first call or has a minimum commitment.**

## Registry

| Provider / API | Used in | Default? | Billing unit | Free tier | Flag |
|---|---|---|---|---|---|
| **Mapbox Temporary Geocoding** (v6 forward, target) | `places/` | yes (geocoding) | request | 100,000 / mo | ✅ |
| **Mapbox Permanent Geocoding** | (forbidden) | no | request | **0 — $5/1k, $5 min** | 🔴 |
| **Mapbox Directions** | `routes/mapbox.go` | fallback | request | 100,000 / mo | ✅ |
| **Google Routes — Compute Routes** | `routes/google.go` | **yes (default routes)** | request | **~5,000 / mo (Pro SKU)** | ⚠️ |
| **Google Place Details** | `places/api.go` | yes | request | **~5,000 / mo (Pro SKU)** | ⚠️ |
| **Google OAuth** (login) | `accounts/` | yes | — | free | ✅ |
| **Google Gemini API** | `trips/` (LLM) | **yes (default LLM)** | tokens / requests | 1,500 req/day, 15 RPM, 1M TPM | ✅* |
| **Anthropic API** | `trips/` (LLM option) | no | tokens | **none** | 🔴 |
| **OpenAI API** (chat + embeddings) | `trips/` LLM + pgvector embeddings | no | tokens | **none** | 🔴 |
| **Ollama** (local LLM / embeddings) | `trips/` (`USE_OLLAMA`) | no | — | free (self-hosted) | ✅ |
| **Amadeus — Flight Offers Search** | `flights/amadeus.go` | **yes (default flights)** | request | ~2,000 / mo | ⚠️ |
| **Amadeus — Hotels / Cities** | `hotels/amadeus.go` | **yes (default hotels)** | request | 200–10,000 / mo (per endpoint) | ⚠️ |
| **IRCTC train search** | `flights/`/trains | — | — | no official API (scrapes irctc.co.in) | 🔴† |
| **DigitalOcean Spaces** | `storage/digitalocean.go` | opt-in (`STORAGE_PROVIDER`) | storage + egress | **$5/mo base, no free tier** | 🔴 |
| **AWS S3** | `storage/s3.go` | opt-in | storage + req | 5 GB for 12 mo only | ⚠️ |
| **Local disk storage** | `storage/local.go` | **yes (default)** | — | free | ✅ |
| **SendGrid** (email) | `notifications/` | opt-in | email | ~100/day trial only | ⚠️ |
| **Twilio** (SMS) | `notifications/` | opt-in | per SMS | **none (trial credit only)** | 🔴 |
| **Firebase Cloud Messaging** (push) | `notifications/` | opt-in | — | free | ✅ |
| **Postgres + pgvector** | `database/` | yes | — | self-hosted | ✅ |

\* **Gemini free tier** is rate-limited (1,500 requests/day) and Google may use
prompts for training. Fine for dev, but the daily cap is a hard ceiling — and if
billing is enabled on the GCP project, calls **silently move to paid**.

† **IRCTC** has no public/free API; hitting `irctc.co.in` directly carries ToS,
legal, and reliability risk independent of cost.

## Current cost exposure (priority order)

1. **Google Routes (default routes provider)** — only ~5,000 free calls/mo since
   Google removed the $200 universal credit on **2025-03-01**. AI trip generation
   calls it per hop. Consider switching `ROUTES_PROVIDER=mapbox` (100k free) or
   caching routes. See [[mapbox-geocoding-v6-decision]].
2. **Google Place Details** — same ~5,000/mo Pro-SKU cap.
3. **Mapbox Permanent Geocoding** — the original overage; eliminated by moving to
   v6 temporary geocoding. Never reintroduce `permanent=true`.
4. **Anthropic / OpenAI** — no free tier; only used if explicitly selected as LLM.
5. **DigitalOcean Spaces / Twilio** — only billed if enabled via env config.

## Guardrails

- Keep cheap/free defaults: `ROUTES_PROVIDER` → prefer `mapbox` or cache;
  `DEFAULT_LLM_PROVIDER=gemini`; `STORAGE_PROVIDER=local` until object storage is
  truly needed.
- Set spending limits + usage alerts on Mapbox, Google Cloud, and Amadeus.
- Cache geocoding, routes, and place details by stable id to stay inside free
  tiers.
- Before adding or switching any third-party API call, follow the check in
  `CLAUDE.md` and update this table.
