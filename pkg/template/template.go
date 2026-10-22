package template

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
)

type Manager interface {
	Render(name string, data any) (string, error)
}

type manager struct {
	tmpl *template.Template
}

func NewManager(fs fs.FS, patterns ...string) (Manager, error) {
	tmpl, err := template.ParseFS(fs, patterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &manager{
		tmpl: tmpl,
	}, nil
}

func (m *manager) Render(name string, data any) (string, error) {
	if m.tmpl.Lookup(name) == nil {
		return "", fmt.Errorf("template not found: %s", name)
	}

	var buf bytes.Buffer
	if err := m.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return buf.String(), nil
}
