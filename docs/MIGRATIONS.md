# Migration Notes

- Migrations live in `backend/migrations`.
- Applied in order:
  1. `001_init.up.sql` (core schema)
  2. `002_seed.up.sql` (sample offers)
- Docker compose includes a `migrate` service that applies both files during startup.

## Rollback (manual)
- Run:
  - `002_seed.down.sql`
  - `001_init.down.sql`
- Use psql against local database.
