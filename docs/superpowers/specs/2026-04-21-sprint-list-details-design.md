# Sprint List Details Design

## Overview

Extend `redmine sprint list <project>` with an opt-in `--details` flag so the command can return full sprint details instead of only the lightweight list payload.

## Problem

The current sprint list response is too sparse for practical use. It only exposes a small subset of each sprint, which makes `table` output narrow and leaves out fields users expect when inspecting sprint history or status.

## Design

### Command Behavior

- `redmine sprint list <project>` remains the entrypoint.
- Add `--details` as a boolean flag, default `false`.
- Without `--details`, the command returns the sprint slice from the list endpoint.
- With `--details`, the command:
  - resolves the project
  - calls `GET /projects/{projectID}/agile_sprints.json`
  - extracts sprint IDs in list order
  - fetches each sprint detail with `GET /projects/{projectID}/agile_sprints/{id}.json`
  - replaces each lightweight sprint with the detailed sprint payload
  - preserves the original list order in the final output

### Field Coverage

When `--details` is enabled, each `Sprint` should contain the full detail set already exposed by the API:

- `id`
- `name`
- `description`
- `status`
- `start_date`
- `end_date`
- `goal`
- `is_default`
- `is_closed`
- `is_archived`

The command should not introduce a second summary mode or a separate detail subcommand.

### Output Shape

- `json` returns `[]Sprint`
- `table` renders the same `[]Sprint` with all exported fields
- `raw` renders the same slice as sanitized JSON

No custom formatter is needed. The shared output layer already converts slices into table rows.

### Data Flow

1. Resolve the project from `<project>`.
2. Fetch sprint IDs from the project sprint list endpoint.
3. If `--details` is disabled, return the lightweight slice immediately.
4. If `--details` is enabled, fetch sprint detail payloads in parallel using the existing batch helper pattern.
5. Merge the detailed payloads back into the original slice order.
6. Write the final sprint slice through the shared output path.

### Error Handling

- If project resolution fails, return the existing project lookup error.
- If the sprint list request fails, return that API error unchanged.
- If any sprint detail request fails, abort the command and return the error for that sprint.
- Do not emit partial results when detail expansion is enabled.

## Implementation Notes

- Update `Sprint` to include `Description`.
- Keep sprint list response decoding compatible with both `agile_sprints` and `sprints` top-level keys.
- Reuse the existing agile client methods instead of adding a new HTTP layer.
- Keep the command under the top-level `sprint` command, with `sprints` as an alias.

## Test Plan

- Root command includes the new `sprint` top-level command.
- `sprint list <project>` resolves the project and returns sprint slices.
- `--details` triggers per-sprint detail fetches.
- The final result preserves the original sprint order.
- Sprint detail decoding includes `description` and other full fields.
- Sprint list decoding accepts the real `sprints` payload shape.
- `table` output receives the expanded sprint fields when `--details` is on.

## Assumptions

- Default behavior stays lightweight for speed.
- `--details` is the only expansion mechanism.
- The list endpoint order is the canonical order and should not be resorted after expansion.
