# Trip Planner — Implementation Roadmap

> Guiding roadmap for completing the app. Written 2026-06-11. This is the source of
> truth for *what to build next and why*; per-feature design details live in their own
> docs (e.g. [AITravelProviders.md](AITravelProviders.md) for the AI provider layer).

## 1. Product Objectives

1. **One app to plan a trip** — destinations, dates, itinerary, stays, transport, budget.
2. **One app to organize the trip while on it** — tickets, contacts, day-of itinerary,
   expenses, documents at your fingertips.
3. **Every trip shape** — a month-long multi-country vacation and a weekend drive must
   both feel native, not like one is a degenerate case of the other.

### Operating constraints

- **Hobby project, run seriously.** Default to free tiers; never *architect around* a free
  tier. Every external service sits behind a provider interface so a paid/standard
  provider (Google Maps, Duffel, Booking.com) is a one-env-var swap if the app takes off.
- **Mapbox first** for geocoding/maps display (existing `places/mapbox.go`, FE map picker);
  **Google Routes** for directions (generous free tier, `routes/google.go` already built,
  Mapbox fallback kept). **Paid-tier users get Google Maps end-to-end** — provider choice
  is per-user via the subscriptions module, not per-deployment.
- **India-first, globally correct.** Primary market is India: UPI settlements, IRCTC/PNR
  realities, permits, festival surge, monsoon seasonality are first-class concerns —
  implemented as providers/templates/data, never hardcoded branches, so the app stays
  usable anywhere.
- **DigitalOcean Spaces** for object storage (already implemented behind the
  `storage.StorageProvider` interface with S3 + local alternatives).
- **Gemini only** for AI trip planning (it does this best today); the `trips/llm_provider.go`
  factory keeps the door open.
- **APIs shaped for great UX.** The existing FE (`../trip-planner-fe`: Dashboard wizard,
  TripDetails tabs, AITripModal, MapPickerModal) is the design target. Endpoints must
  return everything a screen needs in one or two calls (`/trip/:id/complete` is the
  pattern), be additive-only on existing JSON, and never require FE-side joins.

## 2. Current State (what already exists)

| Area | Status | Notes |
|---|---|---|
| Auth: email signup + Google OAuth, JWT | ✅ done | `accounts/` |
| Trip CRUD: plans, hops, days, activities, stays, travellers | ✅ done | `trips/`, full REST surface |
| Itinerary views (`/trip/:id/itinerary`, per-day) | ✅ done | |
| AI generation: `POST /trip/generate`, `/refine`, `/generate/confirm`, city suggestions | ✅ done | Gemini, `trips/ai_controllers.go` |
| Real offers in AI plans: Amadeus flights/hotels, Google/Mapbox routes, mock providers | 🔶 on `ai-travel` branch | See AITravelProviders.md — needs `atlas migrate hash` + live smoke test, then merge |
| Expenses: items, splits, settlements | ✅ done | `expenses/` |
| Documents: upload, DO Spaces/S3/local, `DocumentShare` model | ✅ done | `documents/`, `storage/` |
| Notifications framework: templates, channels, preferences, audit | ✅ built, ❌ not wired to trip events | `notifications/` |
| Subscriptions & feature flags | ✅ done | `subscriptions/`, `featureflags/` |
| Travel knowledge RAG (pgvector + Ollama) | ✅ done | `travelknowledge/` |
| **Publish & public share** | ✅ done | `trips/sharing.go` — publish/unpublish/rotate/public-view/clone; sanitized PII-free projection |
| **Soft delete** | ✅ done | `core.SoftDeleteModel`; TripPlan/Hop/Day/Activity/Stay/Traveller/Document all use it |
| **Trip status state machine** | ✅ done | `trips/lifecycle.go` — typed `TripStatus`, `PATCH /trip/:id/status`, validated transitions |
| **Trip date shifting** | ✅ done | `POST /trip/:id/shift-dates` — atomically shifts plan+hops+days+activities |
| **Today organizer view** | ✅ done | `GET /trip/:id/today` — current hop/stay, check-in/out flags, upcoming transport, emergency contacts |
| **Transport segments** | ✅ done | `trips/transport.go` — TransportMode enum, CRUD under `/trip/:id/transport` and `/transport/:id` |
| **Trip contacts** | ✅ done | `trips/contacts.go` — TripContact model, CRUD under `/trip/:id/contacts` |
| **Pre-trip checklist** | ✅ done | `trips/checklist.go` — ChecklistItem with categories, toggle endpoint, sort order |
| **Password reset** | ✅ done | `accounts/password_reset.go` — hash-stored token, 1h expiry, email enum-safe 200 |
| **Email verification** | ✅ done | `accounts/password_reset.go` — `GET /auth/verify-email`, resend endpoint |
| **EmailProvider interface** | ✅ done | `accounts/email_provider.go` — NoOpEmailProvider logs links for dev; Phase 6 swaps in real provider |
| **Google Calendar integration** | ❌ missing | nothing in the codebase |
| **Weather** | ❌ missing | `TripDay.Weather` field exists, no provider |
| FE: Dashboard, TripDetails, Login/Register, AI modal, map picker | ✅ done | needs new screens per phase below |

## 3. Architecture Principles

1. **Provider factory everywhere.** `places/`, `storage/`, `trips/llm_provider.go`,
   `hotels/`, `flights/`, `routes/` all follow: interface + N implementations + env-selected
   factory + a `mock` implementation for offline dev/CI. Every new external dependency
   (calendar, weather, email, trains) must follow the same shape. Provider selection is
   becoming **per-user**, not only per-deployment: factories grow `GetProviderForUser`
   so subscription tier can pick Google Maps over Mapbox (Phase 7).
2. **Additive API evolution.** The FE ignores unknown JSON fields; never rename or
   re-nest existing fields. New capability = new field or new endpoint.
3. **Mock providers are a release gate.** `*_PROVIDER=mock` must make the full test suite
   pass with no API keys and no network.
4. **Postgres is the source of truth**; provider blobs go to JSONB columns unnormalized
   until we actually query inside them.
5. **Free tier ≠ free-tier-shaped.** Cache aggressively (60s TTL cache exists in
   `core/ttl_cache.go`), rate-limit AI endpoints per user, and degrade gracefully when a
   quota is exhausted — but keep the interfaces provider-grade.

---

## 4. Roadmap

Phases are ordered by user value and dependency. Each phase lists its corner cases —
these are the 90% of real-world travel-organizing scenarios the phase must absorb.

### Phase 1 — Land the AI provider work (in flight, `ai-travel` branch)

Real flights/hotels/routes inside AI-generated plans. Code-complete; remaining:

- [ ] Run `atlas migrate hash` for `20260427000000_add_ai_offer_columns_to_trip_hops.sql`.
- [ ] Live smoke test against Amadeus test env + Google Routes with real keys.
- [ ] Merge to `main`.

Follow-ups already flagged in AITravelProviders.md (geocode tool instead of static city
table, top-N offer heuristics, per-user rate limit on `/trip/generate`) move to Phase 7.

### Phase 2 — Publish & Share (public, no login) ✅ COMPLETE

The single highest-value missing feature: a planned trip becomes a shareable artifact.

**Backend** — all endpoints live in `trips/sharing.go`, registered in `app.go`

```
POST   /api/v1/trip/:id/publish      → generates ShareCode (crypto-random, 10+ chars), sets IsPublic
POST   /api/v1/trip/:id/unpublish    → clears IsPublic (keep ShareCode so re-publish keeps the URL)
POST   /api/v1/trip/:id/share/rotate → new ShareCode (invalidate a leaked link)
GET    /api/v1/public/trip/:share_code  → sanitized read-only trip view (NO auth middleware)
POST   /api/v1/trip/clone/:share_code   → authenticated user copies a public trip as their own
```

- The public endpoint returns a **sanitized projection**, not the raw `TripPlan`:
  itinerary (hops, days, activities), stays (name/dates only, no booking refs or payment
  modes), route summaries, budget *optionally* (owner toggle). It must **never** include:
  travellers' PII (passport numbers, DOB, medical notes, phones), documents, expenses,
  booking references, contact info, or user IDs.
- Public responses are cacheable (`Cache-Control`, ETag) — these get pasted into group
  chats and hit by many logged-out readers.
- Clone copies structure only (hops/days/activities), never documents/expenses/travellers.

**Frontend**: `/t/:shareCode` public route — read-only TripDetails without auth context,
"Open in Trip Planner / Clone this trip" CTA, OG meta tags for link previews.

**Corner cases**

- Owner edits a published trip → public view reflects edits live (it's a view, not a snapshot).
- Owner deletes the trip → public link returns 404, not a 500 or a ghost.
- Share code collision → generate-and-retry on unique index.
- Unpublished/never-published code → identical 404 to nonexistent (don't leak existence).
- A traveller (non-owner) must not be able to publish someone else's trip — publish is
  owner-only; viewing travellers see the publish state.
- Public view of a trip with no dates (see Phase 3 flexible dates) renders day numbers
  ("Day 1") instead of calendar dates.
- Crawler/SEO: public pages should not be indexable by default (`noindex`) — owner opt-in later.

### Phase 3 — Trip Lifecycle & Organizer Mode ("ongoing" trips) ✅ COMPLETE

Make `Status` real and make the app useful *during* travel — objective #2.

**Status state machine** (replaces free-text):

```
planning → confirmed → ongoing → completed
        ↘ cancelled (from any pre-completed state)
```

- Stored as enum; transitions validated server-side (`PATCH /trip/:id/status`).
- **Auto-transitions** via a daily job (and lazily on read): `confirmed → ongoing` when
  `now ∈ [start_date, end_date]` in the *trip's primary timezone*; `ongoing → completed`
  the day after `end_date`. Manual override always wins (a flag records "user-set").
- New `TripPlan.Timezone` (IANA string, derived from first hop's coordinates at save
  time) — without this, "is the trip ongoing?" is wrong by up to a day for users
  planning from another timezone.

**Organizer surface — `GET /trip/:id/today`** (the "day-of" endpoint, one call):

- Today's `TripDay` with activities sorted by time, current hop + stay (with check-in/out
  flags: "you check out today"), next transport leg (from `selected_flight` /
  `route_to_next`), documents tagged to today's segments, emergency contacts.
- FE: when a trip is `ongoing`, TripDetails opens on a **Today** tab; Dashboard pins
  ongoing trips on top.

**Tickets & contacts for ongoing trips** (builds on existing `documents/`):

- Add to `Document`: `doc_type` enum (`ticket`, `visa`, `passport`, `insurance`,
  `booking`, `permit`, `contact`, `other`), optional `trip_hop_id` / `activity_id` /
  `trip_day_id` links, `expires_at`.
- New `TripContact` model: name, role (hotel, guide, driver, embassy, insurance,
  emergency), phone, email, notes, hop link. CRUD under `/trip/:id/contacts`.
- Pre-signed DO Spaces URLs with short TTL for document download; never public.

**Transport segments as first-class data** (today `TripHop.Transportation` is a free-text
string — nowhere to put a train's PNR, platform, or a rental car's pickup window):

- New `TransportSegment` model: mode (`flight | train | bus | car_rental | ferry | taxi`),
  operator, booking ref / PNR, depart & arrive (datetime + timezone + station/terminal),
  cost, hop links. `selected_flight` from Phase 1 becomes one producer of these.
- The Today view and Phase 4 calendar events read transport segments, not the free-text
  field. Trains/buses/ferries get the same treatment flights already have.

**Pre-trip checklist & packing lists** — the single most common organizer behavior:

- `ChecklistItem`: title, category (`booking | documents | packing | money | health`),
  due date, assignee (traveller), done flag. CRUD under `/trip/:id/checklist`.
- Templates seeded by trip type + climate + destination (international → passport/visa/
  forex/insurance; Indian hill stations → permits, warm layers; road trip → FASTag
  recharge, car service).
- Due dates feed Phase 6 notifications ("visa appointment in 3 days").

**Corner cases**

- **Timezones**: trip spanning timezones — "today" is computed against the *current hop's*
  timezone when ongoing, the trip timezone otherwise. Overnight flights: an activity may
  start "yesterday" and land "today"; the today view includes any activity whose
  start *or* end falls today.
- **Flexible dates**: weekend trips often have dates, dream trips don't. Trips without
  `start_date` can never auto-transition; the UI shows relative days. `min_days`/`max_days`
  already support this.
- **Date shifts**: changing `start_date` of a confirmed trip offers to shift all
  hops/days/activities by the same delta (one endpoint:
  `POST /trip/:id/shift-dates {delta_days}`), not silently desync them.
- **Passport/visa expiry**: daily job flags travellers whose passport expires < 6 months
  after trip end, and documents with `expires_at` before trip end → notification.
- **Trip cancelled**: prompt about prepaid stays/booked activities (status →
  `cancelled` on children, never delete — expense records must survive for refunds).
- **Completed trips** become read-mostly: edits allowed (people backfill expenses and
  photos after) but no AI regeneration, no calendar sync.
- **Partially booked reality**: every Stay/Activity keeps `planned | booked | cancelled`
  status independent of trip status — a confirmed trip with zero bookings is normal.
- **India rail reality**: tickets are bought on IRCTC, not in-app. A train segment with
  no PNR yet gets a "Tatkal window opens tomorrow 10:00 IST" reminder; PNR-status checks
  come later (Phase 7). Never assume a bookable API exists.
- **Permits have deadlines**: Inner Line Permits (Sikkim, Arunachal, Ladakh restricted
  areas), forest/park permits — `doc_type=permit` with `expires_at`, plus checklist
  templates auto-adding permit items for destinations that require them.

### Phase 4 — Google Calendar Integration

Auto-create calendar events/reminders from a confirmed trip. **Requires Phase 3** (only
`confirmed`/`ongoing` trips sync; `planning` trips would spam the calendar).

**Design**

- New `calendar/` package, provider interface from day one
  (`CalendarProvider`: `google` | `mock`; ICS export rides along for free — see below).
- **Incremental OAuth**: request `calendar.events` scope only when the user enables sync
  (separate consent from login — users who never sync never see the scope). Store
  refresh token encrypted (AES-GCM, key in env) in a new `accounts.UserIntegration` table.
- Create a dedicated **"Trip Planner" secondary calendar** in the user's account and put
  all events there — one toggle hides everything, deleting our calendar can't touch
  their primary, and we never need to read their existing events.

**Sync model** (the part everyone gets wrong):

- New table `calendar_event_links (entity_type, entity_id, google_event_id, etag, synced_at)`.
- Sync is **one-way push, idempotent, diff-based**: on trip change (or manual "Sync now"),
  compute desired events from the trip, upsert/delete via the link table. Never create
  duplicates on re-sync; deletions in the trip delete the event.
- What becomes an event: hop arrival/departure (transport legs with times), stay check-in
  and check-out (all-day events), activities with `start_time`, and configurable
  reminders (flight −3h, check-out −1h, "trip starts tomorrow").
- Event timezones come from the hop's location, **never** the user's home timezone.
- User edits to our events in Google get overwritten on next sync — documented behavior
  ("the trip is the source of truth"), the alternative (two-way merge) is a tarpit.

```
POST   /api/v1/integrations/google-calendar/connect     → OAuth consent URL
GET    /api/v1/integrations/google-calendar/callback
DELETE /api/v1/integrations/google-calendar              → disconnect + delete our calendar (user choice)
POST   /api/v1/trip/:id/calendar/sync                    → manual sync
PUT    /api/v1/trip/:id/calendar/settings                → enable/disable auto-sync, reminder offsets
GET    /api/v1/trip/:id/calendar.ics                     → ICS file download (no Google needed)
```

**ICS export ships in the same phase** — it's ~a day of work on top of the same event
computation, gives Apple/Outlook users the feature for free, and works for public shared
trips too (objective: organizer value without lock-in).

**Corner cases**

- Token revoked/expired (user revoked in Google settings) → mark integration broken,
  notify, never crash sync; re-consent flow.
- Trip cancelled or unpublished dates → delete all linked events (link table makes this exact).
- All-day vs timed: stays are all-day events (check-in date), flights are timed with
  explicit timezones; activities without times become all-day on their TripDay.
- Multi-traveller trips: only the connecting user's calendar is touched. Other
  travellers connect their own accounts (link table is per-user).
- Quota: Calendar API is free but rate-limited — batch requests, exponential backoff,
  sync debounced (≥60s between auto-syncs per trip).
- Trip deleted while integration disconnected → orphan-cleanup job using the dedicated
  calendar (we can list & delete only our own calendar's events).

### Phase 5 — Collaboration Hardening

The `Traveller` model + invite endpoint exist; finish the permission story.

- **Prerequisite — soft delete everywhere**: add `gorm.DeletedAt` to `core.BaseModel`
  plus a 30-day trash (`GET /trash`, `POST /trash/:id/restore`). Every delete today is
  permanent; in a multi-editor world one misclick kills months of planning. Do this
  *before* handing out editor roles — it touches every delete endpoint, and it only gets
  more expensive to retrofit.
- **Roles**: `organizer` (full control), `editor` (edit itinerary/expenses, not
  publish/delete/travellers), `viewer`. Enforced in middleware per route group — today
  ownership checks are creator-only.
- **Invite flow**: invite by email → if a registered user, in-app notification (the
  notifications framework finally gets wired); else email with join link →
  signup/login → auto-attach `Traveller.UserID`.
- **Corner cases**: traveller leaves mid-planning (their expense splits survive,
  reassigned to "external" participant); organizer deletes account → ownership transfer
  prompt to oldest organizer/editor; children/companions without email are travellers
  with `UserID = null` (already supported — keep it); concurrent edits — last-write-wins
  is fine at this scale, but return `updated_at` and have the FE warn on stale writes;
  a traveller's own PII (passport etc.) is visible only to organizers and themselves.
- **Group decision support**: comments on hops/days/activities and lightweight polls
  ("which weekend?", "beach or hills?"). Without a communication primitive, groups
  coordinate in WhatsApp and only the organizer ever opens the app.

### Phase 6 — Notifications Wired to Trip Events

The framework (`notifications/`) is built; connect producers and one real channel.

- **Email channel** via free tier (Brevo/Resend — pick one, behind the existing channel
  interface). In-app notifications list endpoint for the FE bell icon.
- **Producers**: trip starts tomorrow; passport/visa/document expiry (Phase 3 job);
  invite received; trip shared with you; settlement reminder ("you owe ₹X");
  status auto-transitions ("your trip is now ongoing — here's your Today view").
- **Corner cases**: respect existing `NotificationPreference` (quiet hours, channel
  opt-outs); digest batching for multi-edit bursts (the `NotificationBatch` model
  exists); unsubscribe link in every email (legal requirement, not optional).

**Account hygiene rides on the same email channel** (verified missing from `accounts/`):

- Forgot-password / reset flow (tokened email link) and email verification on signup —
  mandatory the moment a real user registers with email.
- Account deletion (with ownership-transfer prompt for shared trips) and data export
  (JSON dump of trips/expenses/document index) — legal table stakes, cheap to build now.

### Phase 7 — Road-Trip Mode, India-First Features & Planning Polish

**Road-trip mode** (objective #3's "weekend drive" deserves first-class treatment):

- **Search along route**: `GET /routes/poi?along=<route>&category=fuel|food|attraction|ev_charging`
  — Google Places API "search along route" as primary provider (best POI coverage in
  India: dhabas, fuel pumps); Mapbox fallback approximates by category-searching sampled
  points along the polyline (degraded, but the keyless/free path keeps working). Results
  pin onto the trip map and convert to activities/waypoints in one tap.
- **Toll & fuel estimates**: Google Routes API returns toll info (incl. India FASTag
  pricing) — show per-leg toll cost; fuel = distance × vehicle mileage × per-litre price
  (manual user inputs; prices vary by state). Both feed the budget rollup.
- **Waypoints on a leg**: hop-to-hop routes accept intermediate stops (route providers
  already take multi-stop input) with drag-to-reorder.

**Flight & train search**:

- Flights: already live via Amadeus free tier (Phase 1).
- Trains (critical for India): **no official free IRCTC API exists** — design for that.
  New `trains/` provider package: v1 = static timetable search from open data
  (data.gov.in Indian Railways datasets) + deep links to IRCTC/ixigo for booking +
  manual PNR entry on the transport segment; tatkal-window and PNR-status reminders ride
  on Phase 6. European rail: deep links only (no universal free API). The provider
  interface lets a real/paid API drop in later without caller changes.
- Buses: deep links (redBus et al.) on transport segments; no search integration in v1.

**India-first features** (globally harmless, locally killer):

- **UPI settlement links**: expense settlements emit `upi://pay?pa=...&am=...` deep
  links + QR codes — a zero-cost Splitwise-killer. `Traveller` gets an optional
  `upi_id` (PII: visible to organizer + self only).
- **Festival/holiday awareness**: free holiday API (Nager.Date) + travel-knowledge RAG —
  warn at date-picking: "your dates include Diwali week — surge pricing, sold-out
  trains"; or surface the festival as the attraction when that's the point.
- **Season fit**: best-time-to-visit check via RAG when dates are picked (Goa in July is
  a different product than Goa in December; monsoon matters).
- **Formatting & locale**: lakh/crore grouping for INR on the FE; per-country emergency
  number (112 in India) in the Today view; i18n-ready strings now, translations later.

**Planning polish**:

- **Per-tier map provider**: Mapbox default, **Google Maps for paid tiers** — factories
  gain `GetProviderForUser(user)` keyed off the existing `subscriptions` module; the FE
  already abstracts via `MapPickerModal`/`geocodingService`. `map_source` per hop already
  exists, so mixed-source trips work; never migrate stored place IDs between providers.
- **Trip map overview**: all hops + route polylines + stays + route POIs on one map tab —
  the data already exists in `/trip/:id/complete`, mostly FE work.
- **Weather**: Open-Meteo (free, no key) behind a `weather/` provider; fills
  `TripDay.Weather` for trips ≤ 16 days out; refreshed by the daily job; shown in
  Today view and day cards.
- **Activity reordering**: `sort_order` on Activity + batch
  `PUT /days/:id/activities/reorder` — retrofitting order onto existing user data later
  is painful; do it before users accumulate ordered itineraries.
- **Duplicate own trip** ("copy last year's Goa trip"): Phase 2's clone path,
  owner-scoped, no publish required.
- **Budget intelligence**: roll up `actual_spent` from expenses automatically (today the
  fields are parallel and drift); per-hop budget vs actual bars; multi-currency display
  conversion at read time (store original currency always — never convert at write).
- **AI follow-ups from Phase 1**: real geocode tool replacing `staticCityCoords`,
  offer ranking by `(price, rating)`, per-user rate limit on `/trip/generate`,
  optionally the full Gemini tool-loop (design preserved in AITravelProviders.md §TODO).
- **Print/offline**: a print-stylesheet trip view + "download trip as PDF" (the FE
  already has print-layout work; back it with a complete single-call payload —
  `/trip/:id/complete` extended with documents + contacts). Travellers will be offline
  on planes and in Himalayan dead zones; the public share page and PDF are the offline
  answer before considering PWA/service-worker work.
- **Search/filter Dashboard**: by status, date range, tag — trivial backend, big UX win
  once a user has >10 trips.

### Phase 8 — Scale-Up Options (only if it hits off)

Pre-designed swap points and retention features, zero code written now:

- Duffel/Booking.com for bookable offers; paid Gemini tier & tool-loop orchestration;
  S3/CloudFront storage (interface exists); abuse controls on public endpoints; PWA
  offline mode. (Google Maps moved up to Phase 7 — it's a subscription-tier switch now,
  not a scale-up event.)
- **Booking-email forwarding & parsing** (the TripIt feature): forward a confirmation
  email, Gemini parses it into transport segments/stays/documents. Extremely high value,
  needs inbound-email infrastructure — earn it with traction.
- **Post-trip memories**: per-day photos (DO Spaces is ready) + journal notes + travel
  stats (countries, km, spend) — what makes completed trips worth keeping in the app.
- **Visa/vaccination requirement lookup** for country pairs (Indian-passport e-visa/
  visa-on-arrival lists first) via travel-knowledge RAG — answers change often, so it
  must cite sources to be trustworthy.
- **WhatsApp notification channel** (Business API, paid) behind the existing channel
  interface — India's default communication medium, but not free-tier.

---

## 5. Cross-Cutting Corner-Case Checklist

The recurring failure modes of travel apps; every phase above must hold these:

1. **Timezones are per-location, not per-user.** Every datetime attached to a place uses
   that place's timezone. Trip-level "today" derives from the current hop when ongoing.
2. **Dates are optional until they aren't.** Dreaming → planning → confirmed is a
   progression; nothing may crash on a trip with no dates, and date-setting later must
   ripple coherently (shift, don't desync).
3. **Trips mutate after "done".** Expenses get backfilled, docs get added post-trip.
   Completed ≠ frozen.
4. **PII has three audiences** — owner/organizer (everything), co-travellers (itinerary +
   own PII), public (sanitized itinerary only). Every new endpoint declares which
   projection it serves.
5. **External providers fail.** Quota exhausted, key revoked, network down: degrade to
   cached/mock/absent data with a logged warning — never fail the user's primary action
   because an enrichment failed (the Phase 1 enrichment pattern is the template).
6. **Money keeps its original currency.** Convert only at display time; settlements
   computed in the expense's currency; trip totals show mixed-currency honestly.
7. **Deletes cascade visibly.** Deleting a hop orphans days/stays/activities — every
   delete endpoint states what happens to children (block, cascade, or reassign) and the
   FE confirms with specifics.
8. **Idempotency for side-effecting integrations.** Calendar sync, notification sends,
   share-code generation: safe to retry, impossible to duplicate.
9. **Everything works with zero API keys** (mock providers) — CI and new-contributor
   onboarding depend on it.
10. **Booking happens off-app.** Especially in India: IRCTC, redBus, MakeMyTrip own the
    transaction. The app's job is capturing the *result* (PNR, booking ref, times) with
    minimal friction and deep-linking out gracefully — never block a user flow on a
    booking integration we don't have.

## 6. Suggested Sequencing

| Order | Phase | Size | Depends on | Status |
|---|---|---|---|---|
| 1 | Phase 1 — merge `ai-travel` | days | — | 🔶 branch ready, needs atlas hash + smoke test |
| 2 | Phase 2 — publish & share | ~1–2 wks | — | ✅ done |
| 3 | Phase 3 — lifecycle + organizer mode | ~2–3 wks | — | ✅ done |
| 4 | Phase 4 — Google Calendar + ICS | ~2 wks | Phase 3 | ❌ not started |
| 5 | Phase 5 — collaboration roles | ~1–2 wks | — | ❌ not started |
| 6 | Phase 6 — notifications + email channel | ~1 wk | Phases 3, 5 | ❌ not started |
| 7 | Phase 7 — road-trip mode, India-first, polish | ongoing | varies | ❌ not started |
| 8 | Phase 8 — scale-up | deferred | traction | ❌ deferred |

**Implementation notes (2026-06-11)**

- `core.SoftDeleteModel` added; embed it in any model where accidental deletion is catastrophic. User, UserPreferences, notification tables intentionally stay on `BaseModel`.
- `TripPlan.Status` is now typed `*TripStatus` (not `*string`). The DB column is still `text NULL` — no DDL change needed; validation happens in app code.
- Migration `20260611000000_phase2_phase3_account_hygiene.sql` must be applied; run `atlas migrate hash` after applying.
- `EmailProvider` interface in `accounts/email_provider.go` is a no-op for now. Phase 6 swaps in a real provider (Brevo/Resend) via `accounts.SetEmailProvider(...)`.
- All new test files follow the existing convention: DB-dependent tests use `t.Skip`; pure-logic tests run without a DB and cover the 90%+ code paths (token generation, status machine, share code charset/uniqueness, PII sanitization).
- **Architecture change**: `TripHop.Transportation` (free-text) coexists with the new `TransportSegment` table. Phase 3 Today view reads from `transport_segments`. The free-text field is kept for backward compat with AI generation output.
- **Swagger docs** generated at `docs/swagger.json` / `docs/swagger.yaml`; UI served at `/swagger/index.html`. Regenerate with `$(go env GOPATH)/bin/swag init --generalInfo app.go --output docs`. All `json.RawMessage` fields carry `swaggertype:"object"` and all `pq.StringArray` fields carry `swaggertype:"array,string"` — required for swag to parse these types.
- **Development guide** written at `docs/DEVELOPMENT.md` — covers prerequisites, DB setup, migrations, running the server, tests, swagger regeneration, adding migrations, and Docker.

Phases 2 and 3 are independent — 2 first because it's the smallest surface with the
biggest shareable-value payoff, and it exercises the public/sanitized-projection
machinery that 3 and 4 reuse.
