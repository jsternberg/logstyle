package main

import (
	"errors"
	"go/ast"
	"go/types"
	"strings"
)

type Linter interface {
	Inspect(*types.Info, *ast.CallExpr, *types.Func) error
}

var ZapLinter Linter = zapLinter{}

type zapLinter struct{}

func (l zapLinter) Inspect(info *types.Info, expr *ast.CallExpr, fn *types.Func) error {
	if sig, ok := fn.Type().(*types.Signature); ok {
		recv := sig.Recv()
		if recv == nil {
			return nil
		}

		typ := recv.Type()
		for {
			if ptr, ok := typ.(*types.Pointer); ok {
				typ = ptr.Elem()
				continue
			}
			break
		}

		if named, ok := typ.(*types.Named); ok {
			obj := named.Obj()
			if obj.Pkg() == nil {
				return nil
			}

			if stripVendor(obj.Pkg().Path()) == "go.uber.org/zap" && obj.Name() == "Logger" {
				// The receiving type is a zap.Logger instance so we should check
				// the name of the function.
				switch fn.Name() {
				case "Debug", "Info", "Warn", "Error":
					return l.verifyCall(info, expr)
				}
			}
		}
	}
	return nil
}

func (l zapLinter) verifyCall(info *types.Info, expr *ast.CallExpr) error {
	// There should be at least one argument to the call.
	if len(expr.Args) < 1 {
		return nil
	}

	switch arg0 := expr.Args[0].(type) {
	case *ast.BasicLit:
		// This is always ok.
		// TODO(jsternberg): Check if the string is actually good (like newlines).
		return nil
	case *ast.Ident:
		// Use the type checker to ensure that this is a constant.
		obj := info.ObjectOf(arg0)
		if obj != nil {
			if _, ok := obj.(*types.Const); ok {
				// Constants are fine.
				// TODO(jsternberg): Inspect the constant.
				return nil
			}
		}
	}
	return errors.New("call must use a string literal or a constant")
}

func stripVendor(pkgpath string) string {
	// Split the path and search backwards through it to determine if there is a vendor directory.
	paths := strings.Split(pkgpath, "/")
	for i := len(paths) - 1; i >= 0; i-- {
		if paths[i] == "vendor" {
			return strings.Join(paths[i+1:], "/")
		}
	}
	return pkgpath
}
