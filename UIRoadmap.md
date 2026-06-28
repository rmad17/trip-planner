# UI Roadmap & Suggestions

Reviewed: June 2026, against `trip-planner-fe` (Dashboard, TripDetails, TodayView, component set)
and the current backend API surface.

The frontend is in good shape — grouped dashboard with countdown chips and readiness bars,
drag-and-drop destinations, printable itinerary, Today view for ongoing trips. The suggestions
below are ordered roughly by impact.

---

## 1. Collapse Days / Activities / Itinerary into one "Days" view

**Problem:** The sidebar has 13 sections; *Daily Plans*, *Activities*, and *Itinerary* are three
views over the same data (trip days + activities). Users have to decide where to add an activity.

**Suggestion:** A single day-centric view — day cards with activities inline and expandable,
plus a "read/print mode" toggle that replaces the separate Itinerary section. Cuts navigation
by two entries.

**Note:** `TripDetails.js` is ~4,000 lines; this merge is also the natural moment to split it
into per-section components.

## 2. Wire up shipped backend features (currently invisible)

The backend already supports these; the UI doesn't expose them yet:

- **Activity status** — `PATCH /activities/:id/status` accepts `planned / confirmed /
  in_progress / done / completed / skipped / cancelled`. Add a tappable checkmark on each
  activity in Today view and the itinerary. Strikethrough done items, dim skipped ones.
  During an ongoing trip this turns the itinerary into a live checklist.
- **Activity reorder** — `PUT /trip-plans/:id/activities/reorder` (body `{"ids": [...]}`).
  dnd-kit is already used for destinations; same pattern for activities within a day.
  Days (`/days/reorder`) and hops (`/hops/reorder`) have equivalent endpoints.
- **Stay location link** — `stays.location_url` exists. Render an "Open in Maps" button on
  stay cards.

**Effort:** small — an afternoon. Best first pick.

## 3. Trip map

PlaceSearchInput, MapPickerModal, a geocoding service, and place IDs on activities all exist —
but no map is rendered anywhere.

- Overview: hops as numbered pins connected by a route line.
- Dashboard cards: static map thumbnail as fallback when the Wikipedia image misses.

Single most evocative addition for a travel app.

## 4. Budget bar on Overview

Trips have `budget`, expenses are tracked (with BalancesPanel), AI generation produces a budget
breakdown — but nothing shows *spent vs. planned* at a glance. Add a progress bar on Overview
(green → amber → red as actual approaches budget).

## 5. Use AI-generated POIs/restaurants on hops

AI-generated trips fill `pois`, `restaurants`, and `activities` arrays on each hop, but they're
only shown as static text when a day is empty. Render them as chips with a "+" that creates an
activity on a chosen day — one tap from suggestion to plan.

## 6. Styled delete confirmation + undo

Trip delete uses native `window.confirm`, clashing with the polished modals elsewhere. Deletes
are soft-deletes server-side, so an "Undo" toast is feasible and friendlier.

## 7. Smaller polish

- Skeleton cards instead of full-page spinners (stays load per-hop in a waterfall; the page
  pops in piecewise).
- "Copy day as text" button on day cards — travelers paste day plans into group chats; the
  print view already formats this data.
- Custom cover image upload for trips (DO Spaces document storage already exists), Wikipedia
  image as fallback.
- The mobile horizontal section nav scrolls 13 items; the merge in #1 makes it comfortable.

---

## Additional ideas (second wave)

### 8. Calendar export (.ics)

Activities have start/end times and locations. An "Add to calendar" button per day or per trip
(ICS download) puts the itinerary where travelers actually look during the trip. Cheap to build,
no third-party dependency.

### 9. Auto-fill weather

`TripDay.weather` exists as a field but is manual. Wire a free forecast API (e.g. Open-Meteo,
no key required) to populate it for upcoming days using hop coordinates. Show it on day cards
and the Today view.

### 10. Document expiry warnings

Travel documents (passports, visas) are already stored. Surface "passport expires within 6
months of trip end" warnings on the Overview and checklist — a genuinely useful travel
gotcha-catcher.

### 11. Notifications UI

The backend has a `notifications` package that the frontend never surfaces. An in-app bell with
trip reminders (trip starting soon, checklist items due — `checklist_items.due_date` exists)
would close that loop.

### 12. Trip duplication / templates

A clone endpoint already exists for public trips (`POST /trip/clone/:share_code`). Extend to
own trips: "Duplicate trip" with date shifting (the ShiftDatesDialog component already handles
date math). Recurring travelers (annual trips, repeat business routes) get templates for free.

### 13. Collaborative planning

`TripPlan.participants` and traveller invites (`POST /:id/travellers/invite`) exist, and trips
can be shared read-only. Next step is letting invited travellers edit — or, cheaper, per-item
comments so the group can discuss without leaving the app.

### 14. PWA / offline Today view

Travel means flaky connectivity. Caching the Today view and itinerary as a PWA (service worker
+ manifest) means the day plan, contacts, and emergency number work on the metro or abroad
without data. The Today view is exactly the screen that must never fail to load.

### 15. Dark mode

SettingsContext already exists; Tailwind makes this mostly mechanical. Frequent-flyer apps get
used at 6am in dark hotel rooms.

---

## Suggested sequencing

| Order | Item | Why |
|-------|------|-----|
| 1 | #2 Wire up backend features | Smallest effort, makes shipped work visible |
| 2 | #8 ICS export | Small, high traveler value, no dependencies |
| 3 | #1 Days view merge | Biggest structural win; do before adding more sections |
| 4 | #3 Trip map | Flagship visual; benefits from #1's cleanup |
| 5 | #4 Budget bar + #5 POI chips | Quick wins on existing data |
| 6 | #9 Weather auto-fill | Small; needs hop coordinates from #3's geocoding work |
| 7 | #6 Delete confirm + undo, #7 polish items | Fold into whichever section is being touched |
| 8 | #10 Document expiry + #11 Notifications UI | Pair them — expiry warnings are the first notification type |
| 9 | #12 Trip duplication | Reuses clone endpoint + ShiftDatesDialog |
| 10 | #14 PWA / offline Today view | Do after #1 stabilizes the day view it caches |
| 11 | #13 Collaborative planning | Largest scope; needs product decisions (permissions model) |
| 12 | #15 Dark mode | Mechanical; anytime, but lowest urgency |
