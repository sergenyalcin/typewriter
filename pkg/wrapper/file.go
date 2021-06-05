// Copyright 2021 Muvaffak Onus
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wrapper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

func WithImports(im *packages.Imports) FileOption {
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

func NewFile(pkgName, tmpl string, opts ...FileOption) *File {
	f := &File{
		PackageName: pkgName,
		Template:    tmpl,
		Imports:     packages.NewImports(pkgName),
	}
	for _, fn := range opts {
		fn(f)
	}
	return f
}

type File struct {
	HeaderPath  string
	Template    string
	PackageName string
	Imports     *packages.Imports
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
	header, err := ioutil.ReadFile(f.HeaderPath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read header file")
	}
	values := map[string]interface{}{
		"Header":      string(header) + "\n\n// Code generated by typewriter. DO NOT EDIT.",
		"Imports":     importStatements,
		"PackageName": f.PackageName,
	}
	for k, v := range input {
		values[k] = v
	}
	t, err := template.New("file").Parse(f.Template)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, values)
	return []byte(result.String()), errors.Wrap(err, "cannot execute template")
}
