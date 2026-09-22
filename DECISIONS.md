# Decisions

## What did the spec not tell you?

- **Whether weekends count.** Assignments are hours per day over a range, and the seed
  ranges run Monday to Sunday, so I counted every calendar day. That choice decides the
  screen: 180 over-allocated cells in the default window, 54 if weekdays only. Ana is
  56/40 with weekends and exactly 40/40 without. Nothing in the schema marks a working
  calendar. First thing I would confirm.
- **That `from` and `to` select weeks, not days.** The range snaps out to whole ISO weeks,
  matching `date_trunc('week', ...)`. Asking for 2025-12-31 alone returns the whole Dec 29
  week at 56 hours. A day filter needs a prorated denominator for partial weeks and I did
  not invent one.
- **That capacity is flat.** `weekly_hours` applies in every week, partial ones included.
  There is no absence or holiday data to prorate against.
- **The response shape.** A `weeks` header plus one cell per person per week carrying
  `allocated`, `capacity` and `over`. Dense, zeros included, so the grid renders without
  reshaping. The server owns the over rule: strictly greater than, equal is fine.
- **What happens after an edit.** PATCH returns the person and the grid patches capacity
  locally. Changing weekly hours cannot change allocated hours, so a refetch returns
  identical numbers. The cost is not seeing a concurrent edit until the range changes.
- **How much structure to build.** The scaffold is two stub files. I split the API into
  `config`, `models`, `dates`, `security`, `services`, `handlers` and `middleware`, and
  left `main.go` as 60 lines of wiring. Handlers depend on three interfaces rather than on
  a pgx pool, so handler tests run with fakes and no Postgres. Middleware carries panic
  recovery, request logging, security headers and a 1 MiB body cap; validation lives in
  `security` so the rules are in one place, not scattered through handlers. This is more
  layering than 90 minutes of work strictly needs, and I would rather show the seams I
  expect to grow along than a single file I would have to take apart later.
- **What belongs in the repo.** The scaffold ships a three-line `.gitignore`. I took it
  to 45: build output, coverage, env and key files, editor and OS noise. The
  deliberate line is `docker-compose.override.yml`, ignored so a local override cannot
  ride along into the submission while the tracked Compose files stay exactly as
  upstream shipped them.
- **What to test, given no Postgres in the test path.** 21 Go tests across the packages
  plus 7 vitest tests on the date helpers the grid and the API both rely on. Date math is
  tested on both sides because Go and JavaScript disagree about weeks by default. The
  honest gap: `services` has one test, for the over rule, and the SQL itself has none.
- **What to optimize the query for.** Allocation is `O(W log A + K)`: one index probe per
  week against `(start_date, end_date)`, K overlapping rows, O(1) day math per row. The
  people-by-weeks grid is `O(P * W)` and unavoidable for a dense response. At P=500,
  A=126,195, W=3 that is about 7,680 rows touched. W is capped at 52 in validation, which
  bounds the response rather than the scan.

## What did you notice that looked wrong?

- Ana's 56 hours are 15 near-duplicate rows on one project, 14 at 0.5 hours a day and one
  at 1.0. I summed them. Deduping means inventing a rule the schema does not state, and
  they are what makes over-allocation visible at all.
- Eli Nakamura has 0 capacity and 20 allocated hours, so that row is permanently over.
  Left as is, it is a real state.
- 188 of 500 people have no work in the default window, so most of the grid is empty.
  Hiding them is a product call, not a bug fix.
- The planner estimates about 14M rows for the week join against 7,680 actual, which
  triggers JIT: 231 ms in psql, 9.5 ms with `jit=off`. Through the API it is about 125 ms
  for 136 KB, so I left it. A 52-week range is 1.77 MB, which is why it is capped there.

## What did the AI get wrong that you caught?

It wanted to swap the per-week join for a whole-window filter plus a `LATERAL` expansion,
arguing one pass beats W passes. `EXPLAIN ANALYZE` disagreed: on 126,195 assignments
Postgres took the `person_id` index and scanned the table, slower than what it replaced.
The per-week version probes `(start_date, end_date)` once a week and touches about 7,680
rows over three weeks. I kept it and left the measurement in a comment. The argument was
sound in general and wrong about this data with these indexes.

## What would you do differently with a week?

- Settle the weekend question, then make it a working-calendar table with holidays.
- A real day-level range filter, with prorated capacity for partial weeks.
- Integration tests against a real Postgres. The SQL carries the meaning and nothing
  tests it.
- One home for the over rule. The server and the grid both compute it today.
- A version check on PATCH. It is last write wins right now.
- Filter and sort, over-allocated first. Virtualizing made 500 rows scroll, not useful.

Run environment untouched: `docker-compose.yml`, `Makefile`, `api/Dockerfile`,
`db/schema.sql` and `db/seed.sql` are byte-identical to upstream and appear only in the
scaffold commit.
