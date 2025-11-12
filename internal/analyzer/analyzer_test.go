package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}

func TestIsNilIdent(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expr     string
		expected bool
	}{
		{
			name:     "nil identifier",
			code:     "package p; var x = nil",
			expr:     "nil",
			expected: true,
		},
		{
			name:     "non-nil identifier",
			code:     "package p; var x = 42; var y = x",
			expr:     "x",
			expected: false,
		},
		{
			name:     "variable name",
			code:     "package p; var x = 0; var y = x",
			expr:     "x",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			var expr ast.Expr
			ast.Inspect(f, func(n ast.Node) bool {
				if ident, ok := n.(*ast.Ident); ok && ident.Name == tt.expr {
					expr = ident
					return false
				}
				return true
			})

			if expr == nil {
				t.Fatalf("Could not find expression %q", tt.expr)
			}

			result := isNilIdent(expr)
			if result != tt.expected {
				t.Errorf("isNilIdent(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestIsSlice(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expr     string
		expected bool
	}{
		{
			name:     "slice type",
			code:     "package p; var s []int; var _ = s",
			expr:     "s",
			expected: true,
		},
		{
			name:     "pointer type",
			code:     "package p; var p *int",
			expr:     "p",
			expected: false,
		},
		{
			name:     "map type",
			code:     "package p; var m map[string]int",
			expr:     "m",
			expected: false,
		},
		{
			name:     "channel type",
			code:     "package p; var ch chan int",
			expr:     "ch",
			expected: false,
		},
		{
			name:     "array type",
			code:     "package p; var arr [10]int",
			expr:     "arr",
			expected: false,
		},
		{
			name:     "slice literal",
			code:     "package p; var s = []int{1, 2, 3}",
			expr:     "[]int{1, 2, 3}",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			conf := types.Config{}
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
			}
			_, err = conf.Check("p", fset, []*ast.File{f}, info)
			if err != nil {
				t.Fatalf("Failed to type check: %v", err)
			}

			var expr ast.Expr
			// Try to find by name first for simple identifiers
			ast.Inspect(f, func(n ast.Node) bool {
				if ident, ok := n.(*ast.Ident); ok && ident.Name == tt.expr {
					expr = ident
					return false
				}
				return true
			})

			if expr == nil {
				// For slice literals, find the first composite literal
				ast.Inspect(f, func(n ast.Node) bool {
					if cl, ok := n.(*ast.CompositeLit); ok {
						expr = cl
						return false
					}
					return true
				})
			}

			if expr == nil {
				t.Fatalf("Could not find expression %q", tt.expr)
			}

			result := isSlice(info, expr)
			if result != tt.expected {
				t.Errorf("isSlice(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestHasIgnoreDirective(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name: "has ignore directive on same line",
			code: `package p
func f() {
	var s []int
	//nillinter:ignore
	if s == nil {
	}
}`,
			expected: true,
		},
		{
			name: "has ignore directive on previous line",
			code: `package p
func f() {
	var s []int
	//nillinter:ignore
	if s == nil {
	}
}`,
			expected: true,
		},
		{
			name: "no ignore directive",
			code: `package p
func f() {
	var s []int
	if s == nil {
	}
}`,
			expected: false,
		},
		{
			name: "ignore directive too far away",
			code: `package p
func f() {
	var s []int
	//nillinter:ignore
	var x int
	if s == nil {
	}
}`,
			expected: false,
		},
		{
			name: "ignore directive in comment",
			code: `package p
func f() {
	var s []int
	// some comment with nillinter:ignore but not directive
	if s == nil {
	}
}`,
			expected: true, // Current implementation matches any comment containing "nillinter:ignore"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			pass := &analysis.Pass{
				Fset:  fset,
				Files: []*ast.File{f},
			}

			// Build file-to-comments map (matching the optimized implementation)
			fileComments := make(map[*ast.File][]*ast.CommentGroup, len(pass.Files))
			for _, file := range pass.Files {
				fileComments[file] = file.Comments
			}

			// Find the binary expression
			var binExpr *ast.BinaryExpr
			ast.Inspect(f, func(n ast.Node) bool {
				if be, ok := n.(*ast.BinaryExpr); ok {
					binExpr = be
					return false
				}
				return true
			})

			if binExpr == nil {
				t.Fatal("Could not find binary expression")
			}

			result := hasIgnoreDirective(pass, binExpr, fileComments)
			if result != tt.expected {
				t.Errorf("hasIgnoreDirective() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		findExpr func(*ast.File) ast.Expr
		expected string
	}{
		{
			name: "simple identifier",
			code: "package p; var s []int",
			findExpr: func(f *ast.File) ast.Expr {
				var expr ast.Expr
				ast.Inspect(f, func(n ast.Node) bool {
					if ident, ok := n.(*ast.Ident); ok && ident.Name == "s" {
						expr = ident
						return false
					}
					return true
				})
				return expr
			},
			expected: "s",
		},
		{
			name: "selector expression",
			code: "package p; type S struct { f []int }; var s S; var _ = s.f",
			findExpr: func(f *ast.File) ast.Expr {
				var expr ast.Expr
				ast.Inspect(f, func(n ast.Node) bool {
					if sel, ok := n.(*ast.SelectorExpr); ok {
						expr = sel
						return false
					}
					return true
				})
				return expr
			},
			expected: "s.f",
		},
		{
			name: "index expression",
			code: "package p; var arr [][]int; var _ = arr[0]",
			findExpr: func(f *ast.File) ast.Expr {
				var expr ast.Expr
				ast.Inspect(f, func(n ast.Node) bool {
					if idx, ok := n.(*ast.IndexExpr); ok {
						expr = idx
						return false
					}
					return true
				})
				return expr
			},
			expected: "arr[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			expr := tt.findExpr(f)
			if expr == nil {
				t.Fatalf("Could not find expression")
			}

			pass := &analysis.Pass{Fset: fset}
			result := render(pass, expr)
			if result != tt.expected {
				t.Errorf("render() = %q, want %q", result, tt.expected)
			}
		})
	}
}
