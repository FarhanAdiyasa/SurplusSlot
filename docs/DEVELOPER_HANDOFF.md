# Developer Handoff

## Primary Idea
SurplusSlot: merchant surplus offers with timed pickup reservation.

## Backup Idea
EthicalVendor Proofbox (decision recorded, not implemented).

## Core Journey
1. Merchant creates offer.
2. Customer reserves offer (stock decremented atomically).
3. Merchant confirms pickup.

## Critical Files
- Backend server: `backend/internal/httpapi/server.go`
- Data logic: `backend/internal/store/store.go`
- Frontend UI: `frontend/src/app/page.tsx`
- API client: `frontend/src/lib/api.ts`
- Rules: `.cursor/rules/*.mdc`, `AGENTS.md`

## Extension Ideas
- Split merchant and customer UIs with auth.
- Add real payments and cancellation windows.
- Move event `properties` to JSONB and add dashboards.
