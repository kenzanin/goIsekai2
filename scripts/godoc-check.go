//go:build ignore

// Command godoc-check enforces the project godoc rule on every Go package in this module.
//
// It fails when a package has no package comment, or when an exported type, function,
// method, constant or variable has no doc comment that starts with its own name.
// Banner comments are not detected here: they are a review rule, not a lint.
//
// Usage: go run scripts/godoc-check.go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type problem struct {
	pos token.Position
	msg string
}

func main() {
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "godoc-check:", err)
		os.Exit(2)
	}
	problems, err := check(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "godoc-check:", err)
		os.Exit(2)
	}
	for _, p := range problems {
		fmt.Printf("%s:%d: %s\n", p.pos.Filename, p.pos.Line, p.msg)
	}
	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "godoc-check: %d problem(s)\n", len(problems))
		os.Exit(1)
	}
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// check walks root and reports every doc-comment violation, sorted by position.
func check(root string) ([]problem, error) {
	fset := token.NewFileSet()
	var problems []problem
	packages := map[string]bool{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return perr
		}
		dir := filepath.Dir(path)
		packages[dir] = packages[dir] || file.Doc != nil
		problems = append(problems, symbolProblems(fset, file)...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	for dir, documented := range packages {
		if !documented {
			problems = append(problems, problem{
				pos: token.Position{Filename: dir, Line: 1},
				msg: "package has no package comment",
			})
		}
	}
	sort.Slice(problems, func(i, j int) bool {
		if problems[i].pos.Filename != problems[j].pos.Filename {
			return problems[i].pos.Filename < problems[j].pos.Filename
		}
		return problems[i].pos.Line < problems[j].pos.Line
	})
	return problems, nil
}

// symbolProblems reports exported declarations whose doc comment is missing or misplaced.
func symbolProblems(fset *token.FileSet, file *ast.File) []problem {
	var out []problem
	report := func(id *ast.Ident, doc *ast.CommentGroup) {
		if id == nil || !id.IsExported() {
			return
		}
		name := id.Name
		if name == "_" || name == "init" {
			return
		}
		if doc == nil || strings.TrimSpace(doc.Text()) == "" {
			out = append(out, problem{pos: fset.Position(id.Pos()), msg: fmt.Sprintf("%s has no doc comment", name)})
			return
		}
		first := strings.Fields(doc.Text())
		if len(first) == 0 || first[0] != name {
			out = append(out, problem{pos: fset.Position(id.Pos()), msg: fmt.Sprintf("%s doc comment must start with %q", name, name)})
		}
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv != nil && d.Recv.NumFields() == 0 {
				continue
			}
			report(d.Name, d.Doc)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					report(s.Name, docFor(d.Doc, s.Doc))
				case *ast.ValueSpec:
					for _, id := range s.Names {
						report(id, docFor(d.Doc, s.Doc))
					}
				}
			}
		}
	}
	return out
}

// docFor prefers a spec-level doc comment and falls back to the declaration's own.
func docFor(decl, spec *ast.CommentGroup) *ast.CommentGroup {
	if spec != nil {
		return spec
	}
	return decl
}
