-- Reverse of 000001. Every up migration needs a down migration that actually
-- works -- test it before you need it at 2am.

DROP INDEX IF EXISTS idx_todos_due_date;
DROP INDEX IF EXISTS idx_todos_completed_created;
DROP TABLE IF EXISTS todos;
