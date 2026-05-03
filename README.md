# SurplusSlot

SurplusSlot is a side-project MVP for neighborhood cafes to publish end-of-day surplus boxes with timed pickup, letting customers reserve before waste happens.

## Decision Engine (Strict Choice)

### Selected Primary Idea
- **Primary:** SurplusSlot
- **Backup:** EthicalVendor Proofbox

### Weighted Matrix (1-10, higher is better)

Criteria weights:
- Demand confidence: 0.30
- Speed to MVP: 0.25
- Distribution certainty: 0.20
- Monetization speed: 0.15
- Founder fit: 0.10

| Idea | Demand | Speed | Distribution | Monetization | Founder Fit | Weighted Score |
|---|---:|---:|---:|---:|---:|---:|
| SurplusSlot | 9 | 9 | 8 | 8 | 8 | **8.55** |
| EthicalVendor Proofbox | 8 | 7 | 7 | 8 | 8 | 7.55 |

Why SurplusSlot wins: more frequent pain, easier local distribution wedge, and fastest path to measurable transactions.

## 14-Day Build Scope

### Week 1
- Day 1-2: Repo setup, rules, architecture, DB schema, initial API routes.
- Day 3-4: Offer creation and listing.
- Day 5-6: Order flow with stock lock and conflict handling.
- Day 7: Pickup confirmation and event logging.

### Week 2
- Day 8-9: Frontend flow with merchant form, offers list, order + pickup actions.
- Day 10: Seed data + migration automation.
- Day 11: Critical-path tests and lint/typecheck fixes.
- Day 12: API docs and developer handoff docs.
- Day 13: Docker one-command startup, QA pass.
- Day 14: Public launch content + go/no-go metrics review.

## Day-1 Task Checklist
- [x] Create `surplusslot` directory.
- [x] Add AI operating rules (`.cursor/rules/*`, `AGENTS.md`).
- [x] Initialize Next.js frontend and Go backend.
- [x] Define MVP schema (offers, orders, events).
- [x] Implement API error contract.
- [x] Document assumptions.

## Tech Architecture

- **Frontend:** Next.js App Router (`frontend`) for merchant + customer MVP screens.
- **Backend:** Go + chi router (`backend/cmd/api`, `backend/internal/...`) for API and business rules.
- **Database:** Postgres tables for offers, orders, and event logs.
- **Deployment-ready local runtime:** `docker-compose.yml` runs db, migrations, backend, frontend.

Flow:
1. Frontend calls Go API (`NEXT_PUBLIC_API_BASE_URL`).
2. Go API performs validation and transaction-safe writes.
3. Postgres stores domain records and event telemetry.

## Event Tracking Plan (log from day 1)

Log these events via `POST /api/events` (or automatically from API actions):
- `offer_created` with merchant, price, stock, pickup window.
- `offers_viewed` with count and timestamp.
- `order_created` with offerId, quantity, totalCents.
- `order_failed_out_of_stock` with offerId and attempted quantity.
- `pickup_confirmed` with orderId and latency from order time.

Core KPI dashboard queries:
- Offer -> order conversion rate.
- Sell-through by merchant and daypart.
- No-show proxy (placed vs picked_up).
- Median time-to-pickup.

## Launch-in-Public Content Plan (3 posts)

1. **Build Log Post**
   - Headline: "I built a local food-waste MVP in 14 days."
   - CTA: "Reply with your cafe city and I will open pilots there."
2. **Pain-Focused Post**
   - Headline: "Cafes are throwing away margin every night."
   - CTA: "Join SurplusSlot beta: first 10 merchants get free onboarding."
3. **Results Post**
   - Headline: "Week 1 pilot: rescue boxes sold, waste reduced."
   - CTA: "Want this for your neighborhood? Join waitlist."

## Local Setup

### Prerequisites
- Node.js 22+
- PowerShell 5+
- Optional: Docker + Docker Compose (for full stack runtime)

### One-command startup
```bash
docker compose up --build
```

App URLs:
- Frontend: `http://localhost:3000`
- Backend health: `http://localhost:8080/health`

## Development (without Docker)

### Backend
```bash
cd backend
powershell -ExecutionPolicy Bypass -File ../scripts/test-backend.ps1
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

Set env:
- `DATABASE_URL=postgres://postgres:postgres@localhost:5432/surplusslot?sslmode=disable`
- `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`

## Testing

Backend:
```bash
powershell -ExecutionPolicy Bypass -File scripts/test-backend.ps1
```

Frontend:
```bash
cd frontend
npx eslint src --max-warnings=0
npm run build
```

## Verified Commands (Executed)

These commands were executed successfully in this repository:

```powershell
# verify isolated Go toolchain
$env:GOROOT="D:\SIDES\surplusslot\.toolchains\go"
$env:GOPATH="C:\Users\adiya\go"
$env:GOPROXY="https://proxy.golang.org,direct"
$env:GOSUMDB="sum.golang.org"
& "D:\SIDES\surplusslot\.toolchains\go\bin\go.exe" version
& "D:\SIDES\surplusslot\.toolchains\go\bin\go.exe" test ./...

# frontend checks
cd frontend
.\node_modules\.bin\eslint.cmd src --max-warnings=0
npm run build
```

## Troubleshooting

- If backend cannot connect to DB, ensure Postgres is up and migration step completed.
- If frontend cannot call API, check `NEXT_PUBLIC_API_BASE_URL`.
- If migration container fails, remove stale volume with `docker compose down -v` and retry.
- If Docker engine is unavailable, run the non-Docker verification path:
  - `powershell -ExecutionPolicy Bypass -File scripts/test-backend.ps1`
  - `cd frontend; .\node_modules\.bin\eslint.cmd src --max-warnings=0; npm run build`
