package sqldb

import (
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type RawSQLStore struct {
	stmts map[string]string
}

func NewRawStore() *RawSQLStore {
	return &RawSQLStore{stmts: make(map[string]string)}
}

func (s *RawSQLStore) Set(key string, rawStmt string) {
	s.stmts[key] = rawStmt
}

func (s *RawSQLStore) Get(key string) (string, bool) {
	stmt, exists := s.stmts[key]
	return stmt, exists
}

func (s *RawSQLStore) GetAll() map[string]string {
	return s.stmts
}

type StoreGroupedStmtKey struct {
	Group    string
	StmtName string
}

func (k StoreGroupedStmtKey) String() string {
	return k.Group + "." + k.StmtName
}

type GroupFS struct {
	Group string
	FS    embed.FS
}

var RawStoreRegistry []GroupFS

func RegisterGroup(fs embed.FS, group string) {
	RawStoreRegistry = append(RawStoreRegistry, GroupFS{
		FS:    fs,
		Group: group,
	})
}

func LoadRawStmtsToStore(store *RawSQLStore, dbtype string, placeholderPrefix byte) error {
	groupCnt := 0
	stmtCnt := 0
	for _, groupFS := range RawStoreRegistry {
		files, err := groupFS.FS.ReadDir("sql")
		if err != nil {
			return fmt.Errorf("failed to read embedded `sql` dir. %w", err)
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			filename := f.Name()
			ext := filepath.Ext(filename)
			name := strings.TrimSuffix(filename, ext)
			ext = strings.TrimPrefix(ext, ".")
			data, err := groupFS.FS.ReadFile(filepath.Join("sql", filename))
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", filename, err)
			}
			groupedStmtKey := StoreGroupedStmtKey{Group: groupFS.Group, StmtName: name}.String()

			switch ext {
			case dbtype:
				// exact matching file extension -> use it as-is for dialects
				store.Set(groupedStmtKey, string(data))
				stmtCnt++
			case "sql":
				// Standard SQL
				// with Placeholders: `?` (static) and `??` (dynamic)
				if _, exists := store.Get(groupedStmtKey); !exists {
					// Convert static placeholders
					if placeholderPrefix == '?' || placeholderPrefix == 0 {
						// no need to convert
						store.Set(groupedStmtKey, string(data))
					} else {
						store.Set(groupedStmtKey, ReplaceStaticPlaceholders(string(data), placeholderPrefix))
					}
					stmtCnt++
				}
			}
		}
		groupCnt++
	}
	log.Printf("[INFO] %d sql raw stmts loaded for %d groups", stmtCnt, groupCnt)
	return nil
}
