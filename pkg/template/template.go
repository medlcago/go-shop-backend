package template

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"sync"
)

type Manager interface {
	Load(pattern string) error
	LoadFromFS(fs fs.FS, patterns ...string) error
	Render(name string, data any) (string, error)
}

type manager struct {
	tmpl *template.Template
	mu   sync.Mutex
}

func NewManager() *manager {
	return &manager{}
}

func (m *manager) Load(pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tmpl != nil {
		return fmt.Errorf("templates already loaded")
	}

	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	m.tmpl = tmpl
	return nil
}

func (m *manager) LoadFromFS(fs fs.FS, patterns ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tmpl != nil {
		return fmt.Errorf("templates already loaded")
	}

	tmpl, err := template.ParseFS(fs, patterns...)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	m.tmpl = tmpl
	return nil
}

func (m *manager) Render(name string, data any) (string, error) {
	if m.tmpl == nil {
		return "", fmt.Errorf("templates not loaded")
	}

	if m.tmpl.Lookup(name) == nil {
		return "", fmt.Errorf("template not found: %s", name)
	}

	var buf bytes.Buffer
	if err := m.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return buf.String(), nil
}
