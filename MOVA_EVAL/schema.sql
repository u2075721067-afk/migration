PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS snapshots (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  title        TEXT    NOT NULL,
  summary      TEXT,
  tags         TEXT,
  context_text TEXT,
  context_json TEXT,
  source       TEXT,
  author       TEXT
);

CREATE INDEX IF NOT EXISTS idx_snapshots_created_at ON snapshots(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_snapshots_tags ON snapshots(tags);
CREATE INDEX IF NOT EXISTS idx_snapshots_title ON snapshots(title);
