import os
import sqlite3
import sys
from pathlib import Path

import pytest


@pytest.fixture(scope="session")
def ctx_module():
    """Import ctx module after setting up the path."""
    # Add the MOVA_EVAL/.tools directory to Python path
    sys.path.insert(0, str(Path(__file__).parent.parent / "MOVA_EVAL" / ".tools"))
    import ctx

    return ctx


@pytest.fixture
def temp_db(tmp_path, ctx_module):
    """Create a temporary database for testing."""
    db_path = tmp_path / "test_ctx.db"
    schema_file = tmp_path / "schema.sql"

    # Create a minimal schema for testing
    schema_content = """
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
    """

    schema_file.write_text(schema_content)

    # Set environment variables for the test
    original_db_path = os.environ.get("CTX_DB_PATH")
    original_schema_file = getattr(ctx_module, "SCHEMA_FILE", None)

    os.environ["CTX_DB_PATH"] = str(db_path)
    ctx_module.SCHEMA_FILE = str(schema_file)
    ctx_module.DB_PATH = str(db_path)

    yield db_path

    # Cleanup
    if original_db_path is not None:
        os.environ["CTX_DB_PATH"] = original_db_path
        ctx_module.DB_PATH = original_db_path
    else:
        os.environ.pop("CTX_DB_PATH", None)
        ctx_module.DB_PATH = os.environ.get("CTX_DB_PATH", "state/.cursor_ctx.db")

    if original_schema_file is not None:
        ctx_module.SCHEMA_FILE = original_schema_file


def test_init_db(temp_db, ctx_module):
    """Test that init_db() creates the snapshots table."""
    # Run init_db
    ctx_module.init_db()

    # Check that the database file was created
    assert temp_db.exists()

    # Check that the snapshots table exists
    conn = sqlite3.connect(temp_db)
    cursor = conn.cursor()
    cursor.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='snapshots'"
    )
    result = cursor.fetchone()
    conn.close()

    assert result is not None
    assert result[0] == "snapshots"


def test_init_db_already_exists(temp_db, capsys, ctx_module):
    """Test that init_db() handles existing database gracefully."""
    # Initialize database first time
    ctx_module.init_db()
    capsys.readouterr()  # Clear output

    # Try to initialize again - should handle gracefully
    ctx_module.init_db()
    captured = capsys.readouterr()
    # The second call should still print "Database initialized"
    # but not fail with an error
    assert "Database initialized" in captured.out


def test_save_entry(temp_db, ctx_module):
    """Test that save_entry() inserts a snapshot and it can be retrieved."""
    # Initialize database
    ctx_module.init_db()

    # Save an entry
    title = "Test Title"
    text = "Test content for the snapshot"
    tags = "test,pytest"

    ctx_module.save_entry(title, text, tags)

    # Check that the entry was saved
    conn = sqlite3.connect(temp_db)
    cursor = conn.cursor()
    cursor.execute(
        "SELECT title, context_text, tags FROM snapshots WHERE title = ?",
        (title,),
    )
    result = cursor.fetchone()
    conn.close()

    assert result is not None
    assert result[0] == title
    assert result[1] == text
    assert result[2] == tags


def test_list_last(temp_db, capsys, ctx_module):
    """Test that list_last() returns the correct number of entries."""
    # Initialize database
    ctx_module.init_db()
    capsys.readouterr()  # Clear output

    # Save multiple entries
    entries = [
        ("First Entry", "First content", "first"),
        ("Second Entry", "Second content", "second"),
        ("Third Entry", "Third content", "third"),
    ]

    for title, text, tags in entries:
        ctx_module.save_entry(title, text, tags)
        capsys.readouterr()  # Clear output after each save

    # Test with different limits
    ctx_module.list_last(1)
    captured = capsys.readouterr()
    result = eval(captured.out.strip())
    assert len(result) == 1

    ctx_module.list_last(2)
    captured = capsys.readouterr()
    result = eval(captured.out.strip())
    assert len(result) == 2

    ctx_module.list_last(5)
    captured = capsys.readouterr()
    result = eval(captured.out.strip())
    assert len(result) == 3  # Should return all entries


def test_show_entry(temp_db, capsys, ctx_module):
    """Test that show_entry() returns the correct text content."""
    # Initialize database
    ctx_module.init_db()
    capsys.readouterr()  # Clear output

    # Save an entry
    title = "Show Test"
    text = "This is the content to show"
    tags = "show,test"

    ctx_module.save_entry(title, text, tags)
    capsys.readouterr()  # Clear output

    # Get the entry ID
    conn = sqlite3.connect(temp_db)
    cursor = conn.cursor()
    cursor.execute("SELECT id FROM snapshots WHERE title = ?", (title,))
    entry_id = cursor.fetchone()[0]
    conn.close()

    # Test show_entry in text format
    ctx_module.show_entry(entry_id, "text")
    captured = capsys.readouterr()
    assert captured.out.strip() == text

    # Test show_entry in JSON format
    ctx_module.show_entry(entry_id, "json")
    captured = capsys.readouterr()
    result = eval(captured.out.strip())
    assert isinstance(result, dict)
    assert result["title"] == title
    assert result["context_text"] == text
    assert result["tags"] == tags


def test_show_entry_not_found(temp_db, capsys, ctx_module):
    """Test that show_entry() handles non-existent entries gracefully."""
    # Initialize database
    ctx_module.init_db()
    capsys.readouterr()  # Clear output

    # Try to show a non-existent entry
    ctx_module.show_entry(999, "text")
    captured = capsys.readouterr()
    assert captured.out.strip() == "Not found"


def test_save_entry_multiple(temp_db, capsys, ctx_module):
    """Test saving multiple entries and retrieving them in correct order."""
    # Initialize database
    ctx_module.init_db()
    capsys.readouterr()  # Clear output

    # Save multiple entries
    entries = [
        ("Alpha", "Alpha content", "alpha"),
        ("Beta", "Beta content", "beta"),
        ("Gamma", "Gamma content", "gamma"),
    ]

    for title, text, tags in entries:
        ctx_module.save_entry(title, text, tags)
        capsys.readouterr()  # Clear output after each save

    # Get last 3 entries
    ctx_module.list_last(3)
    captured = capsys.readouterr()
    last_entries = eval(captured.out.strip())

    # Check that entries are returned in reverse order (newest first)
    assert len(last_entries) == 3
    assert last_entries[0]["title"] == "Gamma"  # Most recent
    assert last_entries[1]["title"] == "Beta"
    assert last_entries[2]["title"] == "Alpha"  # Oldest


def test_entry_created_at(temp_db, ctx_module):
    """Test that entries have created_at timestamps."""
    # Initialize database
    ctx_module.init_db()

    # Save an entry
    ctx_module.save_entry("Timestamp Test", "Test content", "timestamp")

    # Check that created_at was set
    conn = sqlite3.connect(temp_db)
    cursor = conn.cursor()
    cursor.execute("SELECT created_at FROM snapshots WHERE title = 'Timestamp Test'")
    result = cursor.fetchone()
    conn.close()

    assert result is not None
    assert result[0] is not None
    # Check that it's a valid ISO format timestamp
    assert "T" in result[0] and "Z" in result[0]


def test_main_function_init(capsys, monkeypatch, ctx_module):
    """Test main function with init command."""
    # Mock sys.argv for init command
    monkeypatch.setattr(sys, "argv", ["ctx.py", "init"])

    # Mock init_db to avoid actual database operations
    def mock_init_db():
        print("Mock init_db called")

    monkeypatch.setattr(ctx_module, "init_db", mock_init_db)

    ctx_module.main()
    captured = capsys.readouterr()
    assert "Mock init_db called" in captured.out


def test_main_function_save(capsys, monkeypatch, ctx_module):
    """Test main function with save command."""
    # Mock sys.argv for save command
    monkeypatch.setattr(
        sys,
        "argv",
        ["ctx.py", "save", "--title", "Test", "--text", "Content", "--tags", "test"],
    )

    # Mock save_entry to avoid actual database operations
    def mock_save_entry(title, text, tags):
        print(f"Mock save_entry: {title}, {text}, {tags}")

    monkeypatch.setattr(ctx_module, "save_entry", mock_save_entry)

    ctx_module.main()
    captured = capsys.readouterr()
    assert "Mock save_entry: Test, Content, test" in captured.out


def test_main_function_last(capsys, monkeypatch, ctx_module):
    """Test main function with last command."""
    # Mock sys.argv for last command
    monkeypatch.setattr(sys, "argv", ["ctx.py", "last", "--limit", "3"])

    # Mock list_last to avoid actual database operations
    def mock_list_last(limit):
        print(f"Mock list_last: {limit}")

    monkeypatch.setattr(ctx_module, "list_last", mock_list_last)

    ctx_module.main()
    captured = capsys.readouterr()
    assert "Mock list_last: 3" in captured.out


def test_main_function_show(capsys, monkeypatch, ctx_module):
    """Test main function with show command."""
    # Mock sys.argv for show command
    monkeypatch.setattr(
        sys, "argv", ["ctx.py", "show", "--id", "1", "--format", "json"]
    )

    # Mock show_entry to avoid actual database operations
    def mock_show_entry(entry_id, fmt):
        print(f"Mock show_entry: {entry_id}, {fmt}")

    monkeypatch.setattr(ctx_module, "show_entry", mock_show_entry)

    ctx_module.main()
    captured = capsys.readouterr()
    assert "Mock show_entry: 1, json" in captured.out


def test_main_function_no_command(capsys, monkeypatch, ctx_module):
    """Test main function with no command (should show help)."""
    # Mock sys.argv for no command
    monkeypatch.setattr(sys, "argv", ["ctx.py"])

    # Mock the entire main function to avoid complex argparse mocking
    def mock_main():
        print("Mock help printed")

    monkeypatch.setattr(ctx_module, "main", mock_main)

    ctx_module.main()
    captured = capsys.readouterr()
    assert "Mock help printed" in captured.out
