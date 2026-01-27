package db

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
)

func RunMigrations(conn *sql.DB, dir string) error {
    if err := ensureMigrationsTable(conn); err != nil {
        return err
    }
    applied, err := appliedMigrations(conn)
    if err != nil {
        return err
    }

    entries, err := os.ReadDir(dir)
    if err != nil {
        return err
    }

    for _, e := range entries {
        if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
            continue
        }
        if applied[e.Name()] {
            continue
        }
        path := filepath.Join(dir, e.Name())
        sqlBytes, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        if err := runMigration(conn, e.Name(), string(sqlBytes)); err != nil {
            return err
        }
    }
    return nil
}

func ensureMigrationsTable(conn *sql.DB) error {
    _, err := conn.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            name TEXT PRIMARY KEY,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )
    `)
    return err
}

func appliedMigrations(conn *sql.DB) (map[string]bool, error) {
    rows, err := conn.Query("SELECT name FROM schema_migrations")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    applied := make(map[string]bool)
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            return nil, err
        }
        applied[name] = true
    }
    return applied, rows.Err()
}

func runMigration(conn *sql.DB, name, sqlText string) error {
    tx, err := conn.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if _, err := tx.Exec(sqlText); err != nil {
        return fmt.Errorf("migration %s failed: %w", name, err)
    }
    if _, err := tx.Exec("INSERT INTO schema_migrations (name) VALUES ($1)", name); err != nil {
        return fmt.Errorf("migration %s tracking failed: %w", name, err)
    }
    return tx.Commit()
}
