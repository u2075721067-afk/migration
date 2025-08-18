#!/usr/bin/env python3
"""
Context management utility for Cursor.
"""

import argparse
import datetime
import json
import os
import sqlite3

DB_PATH = os.environ.get("CTX_DB_PATH", "state/.cursor_ctx.db")
SCHEMA_FILE = "schema.sql"


def get_conn():
    return sqlite3.connect(DB_PATH)


def init_db():
    try:
        conn = sqlite3.connect(DB_PATH)
        with open(SCHEMA_FILE, "r", encoding="utf-8") as f:
            schema = f.read()
        conn.executescript(schema)
        conn.commit()
        conn.close()
        print("Database initialized.")
    except sqlite3.OperationalError:
        print("Database already exists or cannot be created.")
        print("Using existing database.")


def save_entry(title, text, tags):
    conn = get_conn()
    created_at = datetime.datetime.utcnow().isoformat() + "Z"
    query = (
        "INSERT INTO snapshots (created_at, title, context_text, tags, summary) "
        "VALUES (?, ?, ?, ?, ?)"
    )
    params = (created_at, title, text, tags, "")
    conn.execute(query, params)
    conn.commit()
    conn.close()
    print("Saved:", title)


def list_last(limit):
    conn = get_conn()
    rows = conn.execute(
        "SELECT id, created_at, title, tags, summary "
        "FROM snapshots ORDER BY id DESC LIMIT ?",
        (limit,),
    ).fetchall()
    conn.close()
    result = []
    for row in rows:
        result.append(
            {
                "id": row[0],
                "created_at": row[1],
                "title": row[2],
                "tags": row[3],
                "summary": row[4],
            }
        )
    print(json.dumps(result, indent=2))


def show_entry(entry_id, fmt):
    conn = get_conn()
    row = conn.execute(
        "SELECT id, created_at, title, tags, summary, context_text "
        "FROM snapshots WHERE id = ?",
        (entry_id,),
    ).fetchone()
    conn.close()
    if not row:
        print("Not found")
        return
    if fmt == "json":
        result = {
            "id": row[0],
            "created_at": row[1],
            "title": row[2],
            "tags": row[3],
            "summary": row[4],
            "context_text": row[5],
        }
        print(json.dumps(result, indent=2))
    else:
        print(row[5])


def main():
    parser = argparse.ArgumentParser(description="Context manager")
    sub = parser.add_subparsers(dest="cmd")

    sub.add_parser("init")

    p_save = sub.add_parser("save")
    p_save.add_argument("--title", required=True)
    p_save.add_argument("--text", required=True)
    p_save.add_argument("--tags", default="")

    p_last = sub.add_parser("last")
    p_last.add_argument("--limit", type=int, default=5)

    p_show = sub.add_parser("show")
    p_show.add_argument("--id", type=int, required=True)
    p_show.add_argument("--format", default="text")

    args = parser.parse_args()

    if args.cmd == "init":
        init_db()
    elif args.cmd == "save":
        save_entry(args.title, args.text, args.tags)
    elif args.cmd == "last":
        list_last(args.limit)
    elif args.cmd == "show":
        show_entry(args.id, args.format)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
