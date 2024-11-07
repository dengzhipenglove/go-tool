
package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
)

// 获取文件中的常量，而且底层类型为 整型和字符串
type ConstIdentItem struct {
	Name        string
	TypeName    string
	Value       int64
	ValueString string
	IsInteger   bool // interger or string
	Comment     string
}

func extract_GoFile_Const_2(filePath string, typeName string) (string, []*ConstIdentItem, error) {
	var pkgName string
	var res = []*ConstIdentItem{}

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	info := types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(".", fset, []*ast.File{astFile}, &info)
	if err != nil {
		return pkgName, nil, err
	}

	pkgName = pkg.Name()

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			vspec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			var typ string
			// get typ
			if vspec.Type == nil && len(vspec.Values) > 0 {
				// "X = 1". With no type but a value. If the constant is untyped,
				// skip this vspec and reset the remembered type.
				typ = ""

				// 获取类型转换的类型名字，例如 const OK = T(2)
				ce, ok := vspec.Values[0].(*ast.CallExpr)
				if ok {
					id, ok := ce.Fun.(*ast.Ident)
					if ok {
						typ = id.Name
					}
				}

			}
			if vspec.Type != nil {
				// "X T". We have a type. Remember it.
				ident, ok := vspec.Type.(*ast.Ident)
				if ok {
					typ = ident.Name
				}

			}
			// 指定了类型，那么类型不匹配不需要
			if typeName != "" && typeName != typ {
				continue
			}

			// extra const data
			for _, name := range vspec.Names {
				if name.Name == "_" {
					continue
				}

				obj, ok := info.Defs[name]
				if !ok {
					panic("fatal: obj not exist" + name.Name)
				}

				comment := ""
				if vspec.Comment != nil {
					comment = vspec.Comment.Text()
				}

				resItem := ConstIdentItem{}

				kst := obj.(*types.Const)

				basic := obj.Type().Underlying().(*types.Basic)
				//fmt.Println(obj.Name(), "-------", kst.Val(), "basicInfo", basic.Info(), "kinfo", basic.Name(), "kind", basic.Kind())

				resItem.Name = name.Name
				resItem.ValueString = kst.Val().ExactString()
				resItem.TypeName = typ

				if basic.Info()&types.IsInteger > 0 {
					resItem.Value, ok = constant.Int64Val(kst.Val())
					resItem.IsInteger = true
				} else if basic.Info()&types.IsString == 0 {
					panic("ident must be interger or string" + name.Name)
				}

				resItem.Comment = comment
				res = append(res, &resItem)
			}
		}
	}
	return pkgName, res, nil
}