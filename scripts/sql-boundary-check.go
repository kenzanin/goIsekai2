// Command sql-boundary-check enforces the project rule that raw SQL lives in
// exactly one module: internal/store/sql. That module may be split across many
// files as it grows. The gate scans Go source for SQL statements in string
// literals, raw-SQL APIs and standalone .sql files outside the module, and
// exits non-zero on any violation.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const boundaryDir = "internal/store/sql"

var sqlStatement = regexp.MustCompile(`(?i)\b(select\s+.+\s+from|insert\s+into|update\s+\S+\s+set|delete\s+from|create\s+(unique\s+)?(table|index|virtual\s+table)|alter\s+table|drop\s+(table|index)|pragma\s+\w+|begin\s+(immediate|exclusive|deferred)|savepoint\s+\w+)\b`)

// rawAPIs are the GORM and database/sql entry points that accept a SQL string.
// They are legal only inside the boundary file.
var rawAPIs = map[string]bool{
	"Raw": true, "Exec": true, "ExecContext": true,
	"Query": true, "QueryContext": true, "QueryRow": true, "QueryRowContext": true,
	"Prepare": true, "PrepareContext": true, "NamedExec": true, "MustExec": true,
}

func main() {
	roots := os.Args[1:]
	if len(roots) == 0 {
		roots = []string{"."}
	}
	var violations []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := d.Name()
			if d.IsDir() && (name == ".git" || name == "vendor" || name == "bin") {
				return filepath.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".sql") {
				return nil
			}
			rel := filepath.ToSlash(filepath.Clean(path))
			if rel == boundaryDir || strings.HasPrefix(rel, boundaryDir+"/") {
				return nil
			}
			if strings.HasSuffix(name, ".sql") {
				violations = append(violations, fmt.Sprintf("%s: .sql file outside %s", rel, boundaryDir))
				return nil
			}
			scanGo(rel, &violations)
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if _, err := os.Stat(boundaryDir); err != nil {
		violations = append(violations, fmt.Sprintf("%s: SQL boundary module missing", boundaryDir))
	}
	if len(violations) > 0 {
		fmt.Fprintln(os.Stderr, "raw SQL found outside "+boundaryDir+":")
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, "  "+v)
		}
		os.Exit(1)
	}
	fmt.Printf("sql-boundary: OK (%s is the only raw-SQL module)\n", boundaryDir)
}

// scanGo reports SQL-bearing literals and raw-SQL calls in one Go file.
func scanGo(path string, violations *[]string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		*violations = append(*violations, fmt.Sprintf("%s: parse error: %v", path, err))
		return
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BasicLit:
			if node.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(node.Value)
			if err != nil {
				return true
			}
			if sqlStatement.MatchString(text) {
				*violations = append(*violations, fmt.Sprintf("%s: SQL in string literal: %s", fset.Position(node.Pos()), abbreviate(text)))
			}
		case *ast.CallExpr:
			sel, ok := node.Fun.(*ast.SelectorExpr)
			if ok && rawAPIs[sel.Sel.Name] {
				*violations = append(*violations, fmt.Sprintf("%s: raw-SQL API %s()", fset.Position(node.Pos()), sel.Sel.Name))
			}
		}
		return true
	})
}

// abbreviate keeps a reported literal to a single readable line.
func abbreviate(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 70 {
		return s[:67] + "..."
	}
	return s
}
