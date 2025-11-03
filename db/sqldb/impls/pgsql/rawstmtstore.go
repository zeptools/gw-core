package pgsql

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/zeptools/gw-core/db/sqldb"
)

var rawStmtStore = sqldb.NewRawStore()

// LoadRawStmtsToStore
// WARNING: Ensure required imports beforehand
func LoadRawStmtsToStore() error {
	groupCnt := 0
	stmtCnt := 0
	for _, groupFS := range sqldb.RawStoreRegistry {
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
			groupedStmtKey := sqldb.StoreGroupedStmtKey{Group: groupFS.Group, StmtName: name}.String()

			switch ext {
			case DBType:
				// exact matching file extension -> use it as-is for dialects
				rawStmtStore.Set(groupedStmtKey, string(data))
				stmtCnt++
			case "sql":
				// Standard SQL
				// with Placeholders: `?` (static) and `??` (dynamic)
				if _, exists := rawStmtStore.Get(groupedStmtKey); !exists {
					// Convert static placeholders
					rawStmtStore.Set(groupedStmtKey, sqldb.ReplaceStaticPlaceholders(string(data), DefaultPlaceholderPrefix))
					stmtCnt++
				}
			}
		}
		groupCnt++
	}
	log.Printf("[INFO][%s] %d sql raw stmts loaded for %d groups", DBType, stmtCnt, groupCnt)
	return nil
}
