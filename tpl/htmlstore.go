package tpl

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

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

func (s *HTMLTemplateStore) LoadBaseTemplates(tplRoot string, fileSuffix string) error {
	return filepath.WalkDir(
		tplRoot,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, fileSuffix) {
				return nil
			}
			rel, _ := filepath.Rel(tplRoot, path)
			key := strings.TrimSuffix(filepath.ToSlash(rel), fileSuffix)

			t := template.New(key)
			t, err = t.ParseFiles(path)
			if err != nil {
				return fmt.Errorf("parse error in %s: %w", path, err)
			}
			s.Base[key] = t
			return nil
		},
	)
}
