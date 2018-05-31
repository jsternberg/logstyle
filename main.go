package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
)

var verbose = flag.Bool("v", false, "verbose output")

func Analyze(w io.Writer, dir string) error {
	fset := token.NewFileSet()
	pkg, err := build.ImportDir(dir, 0)
	if err != nil {
		return err
	}

	pkgs, err := parser.ParseDir(fset, pkg.Dir, func(info os.FileInfo) bool {
		for _, fpath := range pkg.GoFiles {
			if fpath == info.Name() {
				return true
			}
		}
		return false
	}, 0)

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			files = append(files, f)
		}
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{Importer: importer.Default()}
	if _, err := conf.Check(pkg.ImportPath, fset, files, info); err != nil {
		return err
	}

	for _, f := range files {
		ast.Inspect(f, func(node ast.Node) bool {
			switch expr := node.(type) {
			case *ast.CallExpr:
				var ident *ast.Ident
				switch fn := expr.Fun.(type) {
				case *ast.Ident:
					ident = fn
				case *ast.SelectorExpr:
					ident = fn.Sel
				default:
					return false
				}

				obj := info.ObjectOf(ident)
				if obj == nil {
					return false
				}

				// The object should be a *types.Func.
				fn, ok := obj.(*types.Func)
				if !ok {
					return false
				}

				if err := ZapLinter.Inspect(info, expr, fn); err != nil {
					pos := fset.Position(expr.Pos())
					fmt.Fprintf(w, "%s:%d:%d: %s\n", pos.Filename, pos.Line, pos.Column, err)
				}
				return false
			}
			return true
		})
	}
	return nil
}

func realMain() error {
	flag.Parse()

	dir := flag.Arg(0)
	return Analyze(os.Stdout, dir)
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
}
