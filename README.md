# TikTok Carousel Studio

An AI content-to-carousel engine for TikTok creators. You describe a channel in
one sentence; the platform researches the topic, writes the slides, picks a
narrative formula, lays out the design, finds photography, and hands you an
editable canvas and a ZIP of PNGs.

Implementation of `PRODUCT_TECHNICAL_PLAN.md`. Section references (§) below
point back at that document.

---

## Quick start

```bash
make setup      # .env + npm install
make infra      # postgres, redis, minio via docker compose
make api        # terminal 1 — HTTP API on :8080 (migrations run at boot)
make worker     # terminal 2 — generation + export worker
make web        # terminal 3 — Next.js on :3000
```

Open http://localhost:3000, sign in with any email (dev login), and describe a
channel.

**It runs with no API keys.** With `AI_API_KEY` empty a deterministic mock AI
provider takes over, and with `UNSPLASH_ACCESS_KEY` empty a placeholder image
provider does. Every stage — validation, retry, the worker, the editor, the
exporter — runs against the same code paths, so the full flow is exercisable
offline. Add real keys to `.env` for real output; nothing else changes.

```bash
make test          # Go test suite
make smoke         # drives the whole MVP flow against a running stack
make e2e-install   # one-time: download the browser
make e2e           # browser tests for the editor
```

`make smoke` walks the acceptance criteria end to end: login → project → dynamic
form → generation → design → images → autosave with a 409 conflict → export ZIP.

`make e2e` drives a real Chromium against the editor and asserts on **canvas
pixels**. That matters because the editor's failure mode is silent: a style
change can update React state, update the sidebar, and still never reach the
screen. Type checks and unit tests cannot see that; reading the rendered canvas
can. One test also asserts there is exactly one Fabric object per design element
— overlapping async loads once left two, and mutations landed on the hidden
copy.

---

## Architecture

```
Next.js (React, Tailwind, TanStack Query, Zustand, Fabric.js)
                     │ HTTPS, session cookie
                     ▼
              Go API (:8080)                Go worker
     auth · projects · schemas · carousels   research · AI · images · export
     design · assets · export · jobs                 │
                     │                               │
        ┌────────────┴──────────┐          ┌─────────┴──────────┐
        ▼                       ▼          ▼                    ▼
   PostgreSQL                Redis     AI provider          Unsplash
   (source of truth)      (queue,      (OpenAI-compatible)
                           cache,           │
                           rate limit)      ▼
                                          MinIO (private bucket, signed URLs)
```

Two binaries, one wiring (`internal/app`): the API never waits on an AI call,
and the worker never serves HTTP (§82).

### The five layers (§149)

| Layer | Owner | Contract |
|---|---|---|
| 1. Project intelligence | `internal/project` | `Context` — normalized, versioned |
| 2. Content input | `internal/schema` | compiled `Form`, never raw AI JSON |
| 3. Content | `internal/content` | `Content` + semantic gate |
| 4. Design | `internal/design` | Design JSON + layout validator |
| 5. Render | `internal/export`, Fabric adapter | PNG / canvas |

Design JSON is the single source of truth (§2.5). Fabric.js and the Go exporter
are two renderers of the same document, never the other way round.

---

## Decisions this implementation locks in

The blueprint left several choices open. Here is what was chosen and why.

**Export rendering engine — Option B, native Go (§155).** No Node/Chromium
sidecar; the worker stays one Go process. The WYSIWYG risk that comes with
Option B is answered structurally: `internal/fontkit` holds the *only*
implementation of font lookup, text measurement and line wrapping, and it is
used by design validation, by the exporter, and mirrored character-for-character
in `frontend/lib/design/measure.ts`. `internal/export/renderer_test.go` pins the
output with a golden slide; `make golden` refreshes it deliberately.

**Font registry is Vietnamese-gated (§157).** `internal/fontkit/coverage_test.go`
asserts every registered face carries horned vowels and stacked tone marks.
Poppins failed that test and was replaced with Nunito — exactly the failure mode
§157 warned about, caught by a test rather than by a creator.

**Research is pluggable and honest (§16).** `research.SearchProvider` has a real
implementation (set `SEARCH_API_KEY`) and a no-op default. With no search
configured the prompt instructs the model to stay conservative and return *no*
sources rather than inventing URLs, and any non-http URL is stripped before it
reaches `research_sources`.

**Sessions, not JWTs (§162).** Server-side sessions in Postgres behind an
httpOnly cookie. Logout revokes immediately; there is no refresh-token dance.

**Image selection is by id, never by URL (§84).** The browser picks a candidate
id; the server resolves it against that carousel's own cached search results.
The backend never fetches a URL the client supplies — closing an SSRF hole that
the naive "send the candidate object back" design would have opened.

**AI failure never dead-ends the product (§90).** If schema generation fails,
a validated fallback form is used and the creator can still work. If formula
selection fails, `listicle` is used. Only research, content and design failures
fail the job, each with a distinct error code and a user-safe message.

---

## How a carousel is generated

```
POST /projects/:id/carousels/generate   → 202 { carousel_id, job_id }
   ├─ validate input against the stored form   (never trust the browser, §96)
   ├─ INSERT carousel + job in one transaction (§83)
   └─ LPUSH redis
                    worker
   research → content → validate → formula → design → images → finalize
```

Each stage writes a checkpoint (`last_completed_step`, `step_outputs`), so a
retry resumes rather than paying for AI work that already succeeded (§158).
`Idempotency-Key` makes a double submit return the original job (§159).

The client polls `GET /jobs/:id`, which returns the step list with per-step
state — the progress UI renders what the pipeline is actually doing (§15, §102).

### AI is a proposal, never a decision (§145)

Every AI task is typed, validated and audited:

```
generate → parse JSON → typed contract → business validation
                              │                    │
                              └── invalid ─────────┴─→ repair prompt → retry (max 3)
```

* **Schema** — compiled by `internal/schema`: closed component registry, no
  nesting, no `$ref`, capped enums, depth and complexity limits, safe field
  names. An AI-proposed `execute_javascript` component cannot be rendered
  because the frontend registry has no entry for it (§11).
* **Content** — semantic gate, not just valid JSON: hook present, CTA present,
  exact slide count, no duplicate slides, length budgets, and a moderation pass
  that rejects absolute medical/financial claims (§119, §160). Sensitive niches
  get an automatic disclaimer.
* **Design** — fonts must come from the registry, sizes and colours are clamped,
  text is measured and kept inside the safe area, overflow is reported back to
  the model as "rewrite shorter" (§120–§123).

Every call lands in `ai_generations` with task, prompt version, token usage,
latency and status, so a bad carousel can be traced to the generation that
produced it (§70, §92).

---

## Post copy, brand marks and language

**Caption + hashtags.** A carousel is only half a TikTok post. The generation
pipeline drafts a caption and a hashtag set alongside the slides, and the editor
exposes them in a third sidebar tab with a one-click copy. The server records
whether a caption was hand-edited, so regenerating the carousel never silently
overwrites the creator's own words. Hashtags are *suggestions* grounded in the
niche and the carousel's topic — with `SEARCH_API_KEY` set they are also
grounded in live search results. They are not live TikTok trending data, and the
UI says so rather than implying otherwise.

**Brand marks (§39).** A project can carry a `@handle`, a website and an
uploaded logo. Rather than asking the creator to reason about layout, the
defaults encode a recommendation, each overridable:

| Mark | Default | Why |
|---|---|---|
| `@handle` | every slide | a viewer can swipe in anywhere, and it is what converts them into a follower |
| logo | first + last slide | it carries more visual weight, so it belongs at the entry and the exit |
| website | last slide | a conversion element, next to the call to action |

Marks are placed in the outer margin, and the canvas *reserves* that band before
the AI lays anything out — so a brand mark can never collide with generated
copy. `POST /carousels/:id/apply-brand` re-stamps an existing carousel for a
creator who sets up their brand after the fact.

**Two independent languages (§164).**

*UI language* is English/Vietnamese, switched from the header and resolved from a
cookie on the server so the first paint is already correct. Backend errors travel
as stable codes and are localised on the client, so the API never has to know
what language the reader uses. The login screen has no switch: signing in is one
decision, and the switch belongs where the product is actually used.

*Content language* is chosen per project, with **no default** — a wrong default
would write every carousel in that project in the wrong language and only surface
after the first generation. It flows into every AI task: the project brief, the
generated form, research, the slides, the caption and the hashtags.

The offered set is deliberately short and deliberately Latin-script, because the
export renderer draws with the fonts in `internal/fontkit`. A language those
fonts cannot draw would look fine in the browser and export as empty boxes — the
silent failure §157 warns about. `fontkit`'s coverage test asserts every offered
language against every registry font, so the two registries cannot drift apart:

    Vietnamese · English · Indonesian · Malay · Filipino
    Spanish · Portuguese · French · German

## Public pages

`/` and `/pricing` share one chrome with the signed-in app — same palette, same
`Button`/`Card` primitives, same header — so signing in feels like staying in one
place rather than crossing into a different product.

The landing follows the shape a visitor expects: nav with pricing, hero with a
primary and a secondary action, proof, the problem, how it works, three feature
deep-dives with honest mocks of the real surfaces, who it fits, a pricing
preview, FAQ, closing CTA, multi-column footer.

**There is no testimonial or customer-logo strip.** Inventing either would be
fabricated social proof. What the page shows instead is what the product
verifiably does — nine content languages, three canvas formats, eight narrative
formulas, watermark-free export.

`/pricing` shows Free / Pro $20 / Max $100 per month, with yearly billing at 20%
off ($16 and $80 per month, billed $192 and $960 a year), a comparison table and
a billing FAQ. **Billing is not implemented**: nothing on that page is enforced
by the backend, the paid plans carry a disabled "billing not live" control rather
than a checkout, and a banner says so. `frontend/lib/plans.ts` holds the plan
definitions; `e2e/pricing.spec.ts` pins the arithmetic and asserts the page never
offers to take money.

## Editor

`FabricCanvasAdapter` (`frontend/lib/fabric/adapter.ts`) is the only code that
touches Fabric (§98). Coordinates are always logical canvas pixels (1080×1350);
zoom is a viewport concern (§99, §100).

* Undo/redo is an explicit stack of Design JSON snapshots, not browser undo (§31).
* Autosave debounces 1.5s and shows Saving… / Saved ✓ (§32).
* A `PATCH` with a stale version returns **409 plus the server's current
  design**, so a second tab reconciles instead of clobbering (§80).

---

## Layout

```
backend/
  cmd/{server,worker}          two entrypoints, one wiring
  internal/
    app         wiring + router
    auth        sessions, Google OAuth, dev login
    project     identity, AI context, brand kit, schema versions
    schema      AI-schema safety compiler + input validation
    carousel    CRUD, content/design versions, optimistic concurrency
    ai          provider, versioned prompts, typed tasks, retry, audit, mock
    content     content contract + semantic/moderation gate
    design      Design JSON contract + layout validator
    fontkit     THE text measurement engine (shared by validator + exporter)
    image       provider abstraction, Unsplash compliance, search cache
    asset       upload validation, MinIO, signed URLs
    export      native Go renderer, PNG, ZIP, thumbnails
    job         state machine, checkpoints, Redis queue
    worker      generation + export pipelines
    research    search provider abstraction
    registry    fonts + formulas
    ratelimit   per-user daily quotas
  migrations/                  applied automatically at boot

frontend/
  app/          login · onboarding · projects · settings · create · editor
  components/   schema-form · editor · generation · ui
  lib/          api · fabric · design (measure mirror) · editor store
```

---

## Security posture (§84–§87)

Ownership is enforced in SQL, not in handlers — every carousel read joins
`projects` on `user_id`, so swapping an id yields 404, not someone else's data.
The MinIO bucket is private; the browser only ever receives presigned URLs.
Uploads are validated by sniffing the bytes, never the client's extension or
Content-Type. Daily quotas per user: 20 generations, 30 image searches, 20
exports, 100 uploads (§163). Internal errors are logged with a request id and
never serialized to the client.

---

## Not built (deliberately, per §143)

TikTok publishing, Instagram, teams, billing, admin UI, real-time collaboration,
version-restore UI, per-slide regeneration, PDF/JPG export. The architecture
leaves room for each — platform and canvas presets are data (§50), design
versions are already snapshotted (§44), `usage_events` already records the
analytics events (§115) — but none is implemented.

Two things the blueprint asks for that are scaffolded rather than finished:
element locking (§46) exists as a `locked` flag honoured by the editor and the
validator but has no UI, and the prompt evaluation harness (§161) is not built —
prompts are versioned and audited, which is the prerequisite for it.

Brand marks are applied when a carousel is generated. Changing the brand kit does
not retroactively restyle existing carousels; use *Apply brand kit* in the editor
for those.
