// Package index provides full-text search indexing for vault items using SQLite FTS5.
package index

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // pure-Go SQLite driver

	"github.com/zeropass/zeropass/core/vault/types"
)

// Index manages the FTS5 search index backed by SQLite.
type Index struct {
	db   *sql.DB
	path string
}

// Open opens (or creates) the search index at the given path.
func Open(dbPath string) (*Index, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open index db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Index{db: db, path: dbPath}, nil
}

// Close closes the index database.
func (idx *Index) Close() error {
	if idx.db != nil {
		return idx.db.Close()
	}
	return nil
}

// AddToIndex adds an item to the search index.
func (idx *Index) AddToIndex(item *types.Item) error {
	if item == nil {
		return errors.New("item must not be nil")
	}
	// Remove existing entry first.
	_ = idx.removeByID(item.ID)

	res, err := idx.db.Exec(
		`INSERT INTO items_fts (name, type, tags, notes, custom_keys, field_values) VALUES (?, ?, ?, ?, ?, ?)`,
		item.Name,
		string(item.Type),
		strings.Join(item.Tags, " "),
		item.Notes,
		joinMapKeys(item.CustomFields),
		joinMapValues(item.Fields, item.CustomFields),
	)
	if err != nil {
		return fmt.Errorf("index add: %w", err)
	}
	rowid, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get rowid: %w", err)
	}

	_, err = idx.db.Exec(`INSERT OR REPLACE INTO items_map (item_id, fts_rowid) VALUES (?, ?)`, item.ID, rowid)
	if err != nil {
		return fmt.Errorf("index map: %w", err)
	}
	return nil
}

// UpdateIndex updates an item in the search index.
func (idx *Index) UpdateIndex(item *types.Item) error {
	return idx.AddToIndex(item)
}

// RemoveFromIndex removes an item from the search index.
func (idx *Index) RemoveFromIndex(id string) error {
	if id == "" {
		return errors.New("item ID must not be empty")
	}
	return idx.removeByID(id)
}

func (idx *Index) removeByID(id string) error {
	var rowid int64
	err := idx.db.QueryRow(`SELECT fts_rowid FROM items_map WHERE item_id = ?`, id).Scan(&rowid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("lookup rowid: %w", err)
	}
	if _, err := idx.db.Exec(`DELETE FROM items_fts WHERE rowid = ?`, rowid); err != nil {
		return fmt.Errorf("delete fts row: %w", err)
	}
	if _, err := idx.db.Exec(`DELETE FROM items_map WHERE item_id = ?`, id); err != nil {
		return fmt.Errorf("delete map row: %w", err)
	}
	return nil
}

// Search performs a full-text search and returns matching item IDs.
func (idx *Index) Search(query string) ([]string, error) {
	if query == "" {
		return nil, nil
	}

	sanitized := sanitizeQuery(query)

	rows, err := idx.db.Query(
		`SELECT m.item_id FROM items_fts f JOIN items_map m ON f.rowid = m.fts_rowid WHERE items_fts MATCH ? ORDER BY rank`,
		sanitized,
	)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// RebuildIndex clears and rebuilds the index from the given items.
func (idx *Index) RebuildIndex(items []*types.Item) error {
	if _, err := idx.db.Exec(`DROP TABLE IF EXISTS items_fts`); err != nil {
		return fmt.Errorf("drop fts: %w", err)
	}
	if _, err := idx.db.Exec(`DELETE FROM items_map`); err != nil {
		return fmt.Errorf("clear map: %w", err)
	}
	if err := createTables(idx.db); err != nil {
		return err
	}
	for _, item := range items {
		if err := idx.AddToIndex(item); err != nil {
			return err
		}
	}
	return nil
}

// --- helpers ---

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
			name,
			type,
			tags,
			notes,
			custom_keys,
			field_values
		)
	`)
	if err != nil {
		return fmt.Errorf("create FTS table: %w", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS items_map (
			item_id TEXT PRIMARY KEY,
			fts_rowid INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create map table: %w", err)
	}
	return nil
}

func joinMapKeys(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, " ")
}

func joinMapValues(maps ...map[string]string) string {
	var vals []string
	for _, m := range maps {
		for _, v := range m {
			if v != "" {
				vals = append(vals, v)
			}
		}
	}
	return strings.Join(vals, " ")
}

func sanitizeQuery(q string) string {
	// Remove FTS5 special characters to prevent injection.
	replacer := strings.NewReplacer(
		"\"", "",
		"'", "",
		"*", "",
		"(", "",
		")", "",
		":", "",
		"^", "",
		"-", "",
	)
	cleaned := strings.TrimSpace(replacer.Replace(q))
	if cleaned == "" {
		return q
	}

	// Split into terms and add prefix matching.
	terms := strings.Fields(cleaned)
	for i, t := range terms {
		terms[i] = t + "*"
	}
	return strings.Join(terms, " ")
}
