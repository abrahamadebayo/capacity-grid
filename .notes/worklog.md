# Worklog

Running notes on how this got built — decisions, assumptions, dead ends, and anything
left unfinished. Append as you go; a line or two per entry is right.

---

- Allocation = sum(hours_per_day × overlapping calendar days in the ISO week). Weekends count because assignment ranges in the seed include them (e.g. Mon–Sun).
- Capacity per cell = person.weekly_hours (not prorated for partial weeks). Over-allocation is allocated > capacity.
- API shape: `{ weeks: [{start,end}], rows: [{ personId, name, weeklyHours, weeks: [{weekStart, allocated, capacity, over}] }] }`. Nested so the grid can render without reshaping.
- PATCH returns the updated person; the grid patches capacity locally instead of refetching — allocated hours don't change when weekly_hours change.
- Seed has many duplicate assignment rows for the same person/range (Ana especially). Leaving them summed: looks intentional stress for aggregation / over-commit.
- Eli has weekly_hours = 0; any allocation will show as over. Left as-is.
- Docker ran out of disk on first `make up`; pruned ~27GB then stack came up.
- Restructured API into config / models / dates / security / services / handlers / middleware. main.go only wires. Handlers talk to interfaces so tests don't need Postgres.
- Validation lives in security: range ≤ 52 weeks, weeklyHours in [0, 168], DisallowUnknownFields on PATCH, MaxBytesReader 1MiB, security headers middleware.
- Frontend date/range helpers extracted to `dates.ts` with vitest coverage; Go packages have table-driven unit tests (ran via golang container).
- Verified Ana Dec 29 week: API allocated 56 matches SQL sum; over-allocation highlighted in UI. Did not touch Compose/Dockerfile/Makefile/schema/seed.
- Complexity check: seed has ~126k assignments, ~4.7k overlap the default 3-week window. Kept per-week join on `(start_date, end_date)` → O(W log A + K). A whole-window + LATERAL rewrite scanned all A via person_id and was slower/worse for small W. Response is O(P·W) by design (dense grid). Cap W at 52 in security.
- Renamed project folder Untitled → capacity-grid.
- Polish: virtualized the 500-row grid with `@tanstack/react-virtual` (padding rows in tbody). Capacity edit a11y — labelled control, focus return after save/cancel, aria-invalid/live announcements, sr-only over-allocation text, sticky region label.
