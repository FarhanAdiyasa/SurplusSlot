# Assumptions Log

1. **Single city pilot first**
   - Rationale: Allows local pickup behavior learning without multi-region complexity.
   - Impact: No timezone customization in MVP.

2. **Payments are simulated**
   - Rationale: Core risk is reserve/pickup workflow, not payment processing.
   - Impact: `POST /api/orders` creates placed order directly.

3. **Single role UI**
   - Rationale: Speed to MVP in 14 days.
   - Impact: Merchant creation and customer ordering are on one page.

4. **Event properties stored as text**
   - Rationale: Keep schema simple and avoid migration overhead for analytics exploration.
   - Impact: Lightweight event logging; deeper analytics can migrate to JSONB later.
