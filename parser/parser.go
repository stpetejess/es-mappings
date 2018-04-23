package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"strings"
)

const structComment = "es-mapper:json"
const structIgnoreComment = "es-mapper:ignore"

type Parser struct {
	PkgPath     string
	PkgName     string
	StructNames []string
	AllStructs  bool
}

type visitor struct {
	*Parser

	name     string
	explicit bool
}

func hasPrefix(comments string, what string) bool {
	for _, v := range strings.Split(comments, "\n") {
		if strings.HasPrefix(v, what) {
			return true
		}
	}
	return false
}
func (p *Parser) needType(comments string) bool {
	return hasPrefix(comments, structComment)
}
func (p *Parser) ignoreType(comments string) bool {
	return hasPrefix(comments, structIgnoreComment)
}

func (v *visitor) Visit(n ast.Node) (w ast.Visitor) {
	// fmt.Fprintf(os.Stderr, "\nin visit: %v\n", n)
	switch n := n.(type) {
	case *ast.Package:
		return v
	case *ast.File:
		v.PkgName = n.Name.String()
		return v

	case *ast.GenDecl:
		text := n.Doc.Text()
		v.explicit = v.needType(text)
		ignore := v.ignoreType(text)
		// fmt.Fprintf(os.Stderr, "\ngen decl: %v\ntext: %v\nignore: %v\n", v.Parser, text, ignore)
		// second case is to allow ignoring certain structs when using -pkg
		if (!v.explicit && !v.AllStructs) || (v.AllStructs && ignore) {
			return nil
		}
		return v
	case *ast.TypeSpec:
		v.name = n.Name.String()

		// Allow to specify non-structs explicitly independent of '-all' flag.
		if v.explicit {
			v.StructNames = append(v.StructNames, v.name)
			return nil
		}
		return v
	case *ast.StructType:
		v.StructNames = append(v.StructNames, v.name)
		return nil
	}
	return nil
}

func (p *Parser) Parse(fname string, isDir bool) error {
	var err error
	if p.PkgPath, err = getPkgPath(fname, isDir); err != nil {
		return err
	}

	fset := token.NewFileSet()
	if isDir {
		packages, err := parser.ParseDir(fset, fname, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, pckg := range packages {
			ast.Walk(&visitor{Parser: p}, pckg)
		}
	} else {
		f, err := parser.ParseFile(fset, fname, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		ast.Walk(&visitor{Parser: p}, f)
	}
	return nil
}

func getDefaultGoPath() (string, error) {
	output, err := exec.Command("go", "env", "GOPATH").Output()
	return string(output), err
}
