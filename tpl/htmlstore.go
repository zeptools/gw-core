package tpl

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const FileSuffix = ".gohtml"

type HTMLTemplateStore struct {
	Base     map[string]*template.Template // each file → one template
	Combined map[string]*template.Template // composed templates
}

func NewHTMLTemplateStore() *HTMLTemplateStore {
	return &HTMLTemplateStore{
		Base:     make(map[string]*template.Template),
		Combined: make(map[string]*template.Template),
	}
}

func (s *HTMLTemplateStore) LoadBaseTemplates(tplRoot string) error {
	// Normalize the root dir to avoid trailing slash issues
	tplRoot = filepath.Clean(tplRoot)
	err := filepath.WalkDir( // Pre-order Depth-first Traversal
		tplRoot,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := d.Name()
			// Skip Hidden Files & Hidden Directories
			if strings.HasPrefix(name, ".") {
				if d.IsDir() {
					// Hidden Directory
					return fs.SkipDir // skip the whole directory: Do NOT walk into this directory
				}
				// Hidden File
				return nil // skip the file
			}
			if d.IsDir() {
				// Regular Directory
				return nil // just walk into it
			}
			if !strings.HasSuffix(path, FileSuffix) {
				return nil
			}
			// Read file
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// UTF-8 validation
			if !utf8.Valid(data) {
				return fmt.Errorf("file %s is not valid UTF-8", path)
			}
			// compute template key: relative path to the template root without extension
			rel, _ := filepath.Rel(tplRoot, path)
			key := strings.TrimSuffix(filepath.ToSlash(rel), FileSuffix)
			// detect duplicate
			if _, exists := s.Base[key]; exists {
				return fmt.Errorf("duplicate template key detected: %s (file=%s)", key, path)
			}
			// Parse a New Template from the file content
			t := template.New(key)
			t, err = t.Parse(string(data))
			if err != nil {
				return fmt.Errorf("parse error in %s: %w", path, err)
			}
			s.Base[key] = t
			return nil
		},
	)
	if err != nil {
		return err
	}
	// Summary log
	log.Printf("[INFO][TEMPLATE] Loaded %d templates from %s", len(s.Base), tplRoot)
	return nil
}
