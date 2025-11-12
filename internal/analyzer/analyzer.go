package analyzer

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports comparisons of slice values to nil, where an emptiness check
// would be clearer and more correct in many codebases.
//
// It flags:   s == nil   and   s != nil   when s is of slice type.
//
// Suggested fix:
//
//	s == nil  ->  len(s) == 0
//	s != nil  ->  len(s) != 0
//
// Note: This is a style opinion and can change behavior if callers rely on
// distinguishing nil vs empty. To skip linting for specific lines, add the directive:
//
//	//nillinter:ignore
var Analyzer = &analysis.Analyzer{
	Name: "nillinter",
	Doc:  "flag slice comparisons to nil; prefer len(s) == 0 when checking emptiness",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.BinaryExpr)(nil)}

	ins.Preorder(nodeFilter, func(n ast.Node) {
		b, _ := n.(*ast.BinaryExpr)
		if b == nil {
			return
		}
		if b.Op != token.EQL && b.Op != token.NEQ {
			return
		}

		// Check for directive comment above this expression.
		if hasIgnoreDirective(pass, b) {
			return
		}

		// Check the two sides for (slice) == nil or (slice) != nil
		leftIsSlice := isSlice(pass.TypesInfo, b.X)
		rightIsSlice := isSlice(pass.TypesInfo, b.Y)
		leftIsNil := isNilIdent(b.X)
		rightIsNil := isNilIdent(b.Y)

		var sliceExpr ast.Expr
		if leftIsSlice && rightIsNil {
			sliceExpr = b.X
		} else if rightIsSlice && leftIsNil {
			sliceExpr = b.Y
		} else {
			return
		}

		exprStr := render(pass, sliceExpr)
		var replacement string
		if b.Op == token.EQL {
			replacement = "len(" + exprStr + ") == 0"
		} else {
			replacement = "len(" + exprStr + ") != 0"
		}

		msg := "slice compared to nil; use len(s) == 0/!= 0 to test emptiness"
		pass.Report(analysis.Diagnostic{
			Pos:     b.Pos(),
			End:     b.End(),
			Message: msg,
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Replace nil comparison with len(...) == 0/!= 0 check",
					TextEdits: []analysis.TextEdit{
						{Pos: b.Pos(), End: b.End(), NewText: []byte(replacement)},
					},
				},
			},
		})
	})

	return nil, nil
}

func hasIgnoreDirective(pass *analysis.Pass, n ast.Node) bool {
	nodePos := n.Pos()

	// Find the file containing this node
	var file *ast.File
	for _, f := range pass.Files {
		if f.Pos() <= nodePos && nodePos < f.End() {
			file = f
			break
		}
	}
	if file == nil {
		return false
	}

	// Find the comment group immediately preceding this node
	// Comments are associated with the next non-comment token/node
	var closestCommentGroup *ast.CommentGroup
	for _, cg := range file.Comments {
		if cg.Pos() > nodePos {
			break
		}
		closestCommentGroup = cg
	}

	if closestCommentGroup == nil {
		return false
	}

	// Check if the closest comment group contains the ignore directive
	npos := pass.Fset.Position(nodePos)
	for _, c := range closestCommentGroup.List {
		if strings.Contains(c.Text, "nillinter:ignore") {
			cpos := pass.Fset.Position(c.Slash)
			// Comments on the same line or the line immediately before
			if cpos.Filename == npos.Filename && (cpos.Line == npos.Line || cpos.Line == npos.Line-1) {
				return true
			}
		}
	}

	return false
}

func isNilIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "nil"
}

func isSlice(info *types.Info, e ast.Expr) bool {
	t := info.TypeOf(e)
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Slice)
	return ok
}

func render(pass *analysis.Pass, e ast.Expr) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, pass.Fset, e)
	return buf.String()
}
