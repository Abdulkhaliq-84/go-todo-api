-- Migration 000001: create the todos table.
--
-- Plain SQL in numbered files, run by golang-migrate. No Alembic, no
-- autogenerate, no ORM metadata. Your schema history is readable SQL in git,
-- which is less convenient and considerably easier to review.

CREATE TABLE IF NOT EXISTS todos (
    id          UUID PRIMARY KEY,
    title       VARCHAR(200) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    completed   BOOLEAN      NOT NULL DEFAULT FALSE,
    due_date    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- The list endpoint filters on completed and sorts by created_at, so index the
-- pair rather than each column separately.
CREATE INDEX IF NOT EXISTS idx_todos_completed_created
    ON todos (completed, created_at DESC);

-- Partial index: only rows that actually have a due date. Smaller and faster
-- than indexing the whole column when most todos have no deadline.
CREATE INDEX IF NOT EXISTS idx_todos_due_date
    ON todos (due_date)
    WHERE due_date IS NOT NULL;

-- TIMESTAMPTZ, never TIMESTAMP. Postgres stores TIMESTAMPTZ in UTC and converts
-- on the way out; plain TIMESTAMP silently drops the offset and you find out
-- months later when someone's due date is three hours wrong.
