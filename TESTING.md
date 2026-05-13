# Testing

This project ships three layers of automated tests:

- **Fast (unit)** — Go unit tests + Vitest specs. Sub-30-second run.
- **Integration** — Go tests gated by `//go:build integration`, backed by ephemeral Postgres / Redis / MinIO containers via `testcontainers-go`.
- **E2E** — Playwright specs that drive the real frontend against a running docker-compose stack with `APP_ENV=test` helper endpoints mounted.

## Quick commands

```bash
make test-fast    # Go unit + Vitest unit
make test-int     # Go integration (boots testcontainers)
make test-e2e     # Playwright against docker-compose.test.yml
make test-all     # all three in sequence
make cover-html   # generate per-package + HTML coverage reports
```

## Layout

```
server/internal/<pkg>/*_test.go                    # Go unit, no build tag
server/internal/<pkg>/integration_test.go          # Go integration, //go:build integration
server/internal/<pkg>/*_integration_test.go        # alternative naming, same tag
server/test/containers/                            # testcontainers helpers
server/test/auth/token.go                          # MustIssueToken helper for tests
server/test/fixtures/                              # SQL fixtures + Load/Truncate helpers

client/src/**/*.spec.ts                            # Vitest specs co-located

e2e/*.spec.ts                                      # Playwright specs
e2e/helpers/reset.ts                               # /api/test/reset, seed-song, promote
e2e/helpers/db.ts                                  # direct pg client for invariant checks
e2e/fixtures/                                      # binary fixtures (PNGs, etc.)
```

## Naming conventions

- **Go unit:** `TestType_Method_Behavior` (e.g. `TestJWTService_ValidateAccessToken_ExpiredTokenReturnsErrExpired`).
- **Go integration:** same naming, with an `_INT` suffix to enable easy filtering via `-run "_INT$"` (e.g. `TestSongService_EnqueueSong_SecondEnqueueIsDeduped_INT`). File must start with `//go:build integration` as line 1.
- **Vitest:** `describe('Subject', () => { it('does X when Y', ...) })`. Files end in `.spec.ts` and live next to the code under test.
- **Playwright:** `test('user action description', ...)` inside `test.describe('Flow name', ...)`. Files end in `.spec.ts` in `e2e/`.

## Coverage targets (informational, not enforced)

Backend gates are documented as targets, verified manually via `make cover-html` and `go tool cover -func=coverage.out`. Numbers are realistic for Go's branch-per-`if err != nil` pattern and the project's fx/DI boilerplate — chasing high coverage there pushes tests into trivial error wrappers without adding meaningful signal.

| Package | Target | Current* |
|---|---|---|
| `internal/auth` | 45% | **77.9%** |
| `internal/utils/pagination` | 60% | **100.0%** |
| `internal/song` | 35% | 40.1% |
| `internal/files` | 45% | 50.9% |
| `internal/role` | 40% | 46.6% |
| `internal/fingerprint` | 55% | 60.8% |
| `internal/user` | 5% (mock-only scope) | 5.8% |

\* as of 2026-05-13 — re-run `make cover-html` to refresh.

Frontend coverage is **enforced** by `client/vitest.config.ts`:

```ts
thresholds: { lines: 50, functions: 50, branches: 40, statements: 50 }
```

`npm run test:coverage` fails the run if any threshold is missed.

## Test infrastructure

### Reset / seed (E2E)

The server registers three `/api/test/*` endpoints only when `APP_ENV=test`:

| Endpoint | Body | Effect |
|---|---|---|
| `POST /api/test/reset` | — | Truncates `fingerprints, songs, user_roles, users, files`; re-seeds `user` and `admin` roles. |
| `POST /api/test/seed-song` | `{title, artist, duration?, source_id?, uploaded_by?}` | Inserts a song row directly, bypassing the worker pipeline. |
| `POST /api/test/promote` | `{email_hash, role}` | Adds a role to a user, identified by email_hash. |

These endpoints are absent from production builds — guarded by a runtime `APP_ENV=test` check in `server/internal/app/testendpoints.go`.

### Container lifecycle (integration)

`server/test/containers/{postgres,redis,minio}.go` provides `Start*` helpers that:
- Boot the container via `testcontainers-go`.
- Register `t.Cleanup` to terminate it after the test (panics included).
- For Postgres, run every goose migration in `server/migrations/` against the fresh DB before returning.

### Mocked external dependencies

- **Spotify** (`SongMetadataSource`) and **yt-dlp** (`SongDownloader`) are mocked at the interface boundary inside Go unit and integration tests via `testify/mock`. There is no separate mock-upstream service.
- **S3 / MinIO** is mocked at the `S3Client` interface boundary in unit tests; integration tests use a real MinIO container.
- **Redis / asynq** is exercised against a real Redis container in integration tests.

The "third-party API outage" lab requirement (4.5) is covered by `TestSongService_AddSong_MetadataSourceOutageReturnsErrorNoOrphan_INT` in `server/internal/song/integration_test.go`, which simulates a Spotify outage via the mocked interface and verifies the worker fails cleanly without orphan rows.

## Adding a new test

| Want to test… | Do this |
|---|---|
| A service method in isolation | New `*_test.go` next to the file. Use `testify/mock` to fake collaborators. No build tag. |
| End-to-end against a real DB / Redis / S3 | New `*_test.go` or `*_integration_test.go` with `//go:build integration` as line 1. Use `server/test/containers/*` helpers. |
| A Vue component or composable | New `*.spec.ts` next to the component. Mount with `@vue/test-utils`. |
| A user flow through the browser | New `e2e/<flow>.spec.ts`. Call `resetServer()` in `beforeEach`. |
