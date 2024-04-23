package precompile

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

type Handle struct {
	address     string
	structName  string
	packageName string
}

func NewHandle(address, structName, packageName string) Handle {
	return Handle{
		address:     address,
		structName:  structName,
		packageName: packageName,
	}

}

func (h Handle) addImport(file *ast.File) bool {
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(h.packageName),
		},
	}
	var importDecl *ast.GenDecl
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if ok && genDecl.Tok == token.IMPORT {
			importDecl = genDecl
			break
		}
	}
	if importDecl == nil {
		importDecl = &ast.GenDecl{
			Tok: token.IMPORT,
		}
		file.Decls = append([]ast.Decl{importDecl}, file.Decls...)
	}
	importDecl.Specs = append(importDecl.Specs, importSpec)
	return true
}

func (h Handle) addMapItem(node ast.Node) bool {
	// Check if the node is a ValueSpec (which includes var declarations)
	valueSpec, ok := node.(*ast.ValueSpec)
	if !ok {
		return true // Continue traversing the AST
	}

	// Assuming the map variable is named "myMap"
	for idx, name := range valueSpec.Names {
		if name.Name == "PrecompiledContractsCancun" {
			// Create a new map item
			newItem := &ast.KeyValueExpr{
				Key:   &ast.BasicLit{Value: "common.BytesToAddress([]byte{" + h.address + "})"},
				Value: &ast.BasicLit{Value: "&" + h.packageName + "." + h.structName + "{}"},
			}

			// Add the new item to the map variable
			value := valueSpec.Values[idx].(*ast.CompositeLit)
			value.Elts = append(value.Elts, newItem)
			valueSpec.Values[idx] = value
		}
	}

	return true // Continue traversing the AST
}

func (h Handle) registerPrecompile(pathToOpGet string) error {
	fileSet := token.NewFileSet()
	contractPath := pathToOpGet + "/core/vm/contracts.go"
	// check if file not exists return error
	file, err := parser.ParseFile(fileSet, contractPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Edit the AST
	ast.Inspect(file, h.addMapItem)
	h.addImport(file)

	// Write the modified AST back to the file
	fileOut, err := os.Create(contractPath)
	if err != nil {
		return err
	}
	defer fileOut.Close()

	if err := format.Node(fileOut, fileSet, file); err != nil {
		return err
	}

	return nil
}
