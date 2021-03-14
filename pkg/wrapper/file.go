package wrapper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/muvaf/typewriter/pkg/imports"

	"github.com/pkg/errors"
)

func WithImports(im *imports.Map) FileOption {
	return func(f *File) {
		f.Imports = im
	}
}

func WithHeaderPath(h string) FileOption {
	return func(f *File) {
		f.HeaderPath = h
	}
}

type FileOption func(*File)

func NewFile(pkg, tmplPath string, opts ...FileOption) *File {
	f := &File{
		Package:      pkg,
		TemplatePath: tmplPath,
		Imports:      imports.NewMap(pkg),
	}
	for _, fn := range opts {
		fn(f)
	}
	return f
}

type File struct {
	HeaderPath   string
	TemplatePath string
	Package      string
	Imports      *imports.Map
}

// Wrap writes the objects to the file one by one.
func (f *File) Wrap(input map[string]interface{}) ([]byte, error) {
	importStatements := ""
	for p, a := range f.Imports.Imports {
		// We always use an alias because package name does not necessarily equal
		// to that the last word in the path, hence it's not completely safe to
		// not use an alias even though there is no conflict.
		importStatements += fmt.Sprintf("%s \"%s\"\n", a, p)
	}
	values := map[string]interface{}{
		"Header":  "// Code generated by muvaf/typewriter. DO NOT EDIT.",
		"Imports": importStatements,
		"Package": f.Package,
	}
	for k, v := range input {
		values[k] = v
	}
	tmpl, err := ioutil.ReadFile(f.TemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read template file")
	}
	tpl := string(tmpl)
	t, err := template.New("file").Parse(tpl)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, values)
	return result.Bytes(), errors.Wrap(err, "cannot execute template")
}
