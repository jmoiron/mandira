package mandira

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Loader struct {
	Path    string
	Preload bool
	Loaded  bool
	cache   map[string]*Template
}

func anysuffix(has string, any ...string) bool {
	for _, s := range any {
		if strings.HasSuffix(has, s) {
			return true
		}
	}
	return false
}

func IsTemplate(path string) bool {
	return anysuffix(path, "mnd", "mandira", "mda")
}

func NewLoader(path string, preload bool) *Loader {
	loader := &Loader{Path: path, Preload: preload}
	loader.cache = map[string]*Template{}
	if preload {
		loader.Refresh()
	}
	return loader
}

func (l *Loader) visitor(path string, f os.FileInfo, err error) error {
	if err != nil || f == nil {
		return nil
	}
	if f.Mode().IsRegular() && IsTemplate(path) {
		tpl, err := ParseFile(path)
		if err != nil {
			return err
		}
		l.cache[strings.TrimPrefix(path, l.Path)] = tpl
	}
	return nil
}

func (l *Loader) Refresh() error {
	err := filepath.Walk(l.Path, l.visitor)
	l.Loaded = true
	return err
}

func (l *Loader) Get(path string) (*Template, error) {
	var err error
	if l.Preload && !l.Loaded {
		err = l.Refresh()
		if err != nil {
			return nil, err
		}
	}

	if l.Preload {
		tpl, ok := l.cache[path]
		if !ok {
			return nil, errors.New("Template " + path + " does not exist or cannot be loaded.")
		}
		return tpl, nil
	}

	return ParseFile(filepath.Join(l.Path, path))
}

func (l *Loader) MustGet(path string) *Template {
	t, err := l.Get(path)
	if err != nil {
		panic(err)
	}
	return t
}

// Return the internal cache
func (l *Loader) Cache() map[string]*Template {
	return l.cache
}

// If you want to add a template sourced from elsewhere to the loader, you
// can do it here and continue to use the loader.
func (l *Loader) Add(path string, template *Template) {
	l.cache[path] = template
}
