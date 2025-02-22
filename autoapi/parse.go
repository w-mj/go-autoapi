package autoapi

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func WalkPath(root string) {
	var err error
	root, err = filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.Name() == "autoapi_gen" {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
			p := &Parser{RootDir: root}
			p.Init()
			p.ParseFile(path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

type FuncInfo struct {
	Name    string
	NumArgs int
}

type Parser struct {
	RootDir string
	Base    string
	Name    string
	Package string

	ImportName map[string]string
	ImportPath map[string]string
	UsedImport map[string]string // path: name or path: nil

	HandlerFunc  []ast.Decl
	FuncInfoList []FuncInfo
}

func (p *Parser) Init() {
	p.ImportName = make(map[string]string)
	p.ImportPath = make(map[string]string)
	p.UsedImport = make(map[string]string)
}

func (p *Parser) ParseFile(root string) {
	root = filepath.ToSlash(root)
	p.Base, p.Name = path.Split(root)
	p.Base = filepath.ToSlash(p.Base)
	p.RootDir = filepath.ToSlash(p.RootDir)
	var err error
	a := token.NewFileSet()
	pack, err := parser.ParseFile(a, root, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Parse %s error: %s\n", root, err.Error())
		return
	}
	p.Package = pack.Name.Name
	ast.Inspect(pack, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.ImportSpec:
			p.ParseImportSpec(n.(*ast.ImportSpec))
		case *ast.FuncDecl:
			p.ParseFunction(n.(*ast.FuncDecl))
		}
		return true
	})

	if len(p.HandlerFunc) == 0 {
		return
	}
	p.GenerateAddRouterFunc()
	var im []*ast.ImportSpec
	im = append(im, &ast.ImportSpec{
		Path: &ast.BasicLit{Value: "\"github.com/gin-gonic/gin\"", Kind: token.STRING},
	})
	var imp []ast.Spec
	for _, d := range im {
		imp = append(imp, d)
	}
	imdec := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: imp,
	}

	out := &ast.File{
		Name:    &ast.Ident{Name: p.Package},
		Imports: im,
		Decls:   append([]ast.Decl{imdec}, p.HandlerFunc...),
	}
	outfset := token.NewFileSet()
	outDir := p.Base
	_ = os.MkdirAll(outDir, 0644)
	name := strings.TrimSuffix(p.Name, ".go") + "_gen.go"
	genFile, err := os.OpenFile(path.Join(outDir, name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("cannot open %s, error: %s", path.Join(outDir, p.Name), err.Error())
		return
	}
	if err := format.Node(genFile, outfset, out); err != nil {
		fmt.Printf("cannot write to file, error: %s", err.Error())
	}
	_ = genFile.Close()
	jsName := strings.TrimSuffix(p.Name, ".go") + ".js"
	dir := path.Join(p.RootDir, "api_gen", strings.TrimPrefix(p.Base, p.RootDir))
	os.MkdirAll(dir, 0755)
	jsFile, err := os.OpenFile(path.Join(dir, jsName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("cannot open generated js file, error: %s\n", err.Error())
		return
	}
	p.GenerateJavaScript(jsFile)
	_ = jsFile.Close()
}

func (p *Parser) ParseImportSpec(im *ast.ImportSpec) {
	var na string
	if im.Name != nil {
		na = im.Name.Name
	} else {
		id := strings.LastIndexByte(im.Path.Value, '/')
		if id < 0 {
			na = im.Path.Value
		} else {
			na = im.Path.Value[id+1:]
		}
	}
	p.ImportPath[im.Path.Value] = im.Path.Value
	p.ImportName[na] = im.Path.Value
}

func (p *Parser) ParseFunction(fn *ast.FuncDecl) {
	if fn.Doc == nil || !strings.Contains(strings.ToLower(fn.Doc.Text()), "@autoapi") {
		return
	}
	han := &ast.FuncDecl{}
	//han.Doc = &ast.CommentGroup{List: []*ast.Comment{{
	//	Text: fmt.Sprintf("// Handler for %s, generated by AutoAPI\n", fn.Name.Name),
	//}}}
	han.Name = &ast.Ident{Name: fmt.Sprintf("AutoAPI_handler_%s", fn.Name.Name)}
	han.Type = &ast.FuncType{
		Params: &ast.FieldList{List: []*ast.Field{{
			Names: []*ast.Ident{{Name: "c"}},
			Type:  &ast.StarExpr{X: &ast.SelectorExpr{Sel: &ast.Ident{Name: "Context"}, X: &ast.Ident{Name: "gin"}}},
		}}},
	}
	han.Body = &ast.BlockStmt{}

	if fn.Type.Params.NumFields() > 2 {
		fmt.Printf("Handler function can only be 0, 1 or 2 params, but %s has %d.", fn.Name.Name, fn.Type.Params.NumFields())
		return
	}
	if fn.Type.Results.NumFields() > 2 {
		fmt.Printf("Handler function can only be 0, 1 or 2 results, but %s has %d.", fn.Name.Name, fn.Type.Params.NumFields())
		return
	}
	han.Body.List = p.PrepareArguments(fn, han.Body.List)
	han.Body.List = p.PrepareCall(fn, han.Body.List)
	han.Body.List = p.PrepareReturns(fn, han.Body.List)
	p.HandlerFunc = append(p.HandlerFunc, han)
	p.FuncInfoList = append(p.FuncInfoList, FuncInfo{Name: fn.Name.Name, NumArgs: fn.Type.Params.NumFields()})

	fmt.Printf("\tfunc %s(", fn.Name.Name)
	for _, t := range fn.Type.Params.List {
		fmt.Printf("%v,", t.Type)
	}
	fmt.Printf(") -> ")
	if fn.Type.Results != nil {
		for _, t := range fn.Type.Results.List {
			fmt.Printf("%v,", t.Type)
		}
	}

	fmt.Printf("\n")
}

func (p *Parser) PrepareArguments(fn *ast.FuncDecl, List []ast.Stmt) []ast.Stmt {
	if fn.Type.Params.NumFields() > 0 {
		List = p.PrepareOneArgument(fn, fn.Type.Params.List[0], 1, List)
		if fn.Type.Params.NumFields() == 2 {
			List = p.PrepareOneArgument(fn, fn.Type.Params.List[1], 2, List)
		}
	}

	if fn.Type.Results.NumFields() == 1 {
		List = append(List, &ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR,
			Specs: []ast.Spec{&ast.ValueSpec{
				Names: []*ast.Ident{{Name: "r1"}},
				Type:  ast.NewIdent("any"),
			}},
		}})
	} else if fn.Type.Results.NumFields() == 2 {
		List = append(List, &ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR,
			Specs: []ast.Spec{&ast.ValueSpec{
				Names: []*ast.Ident{{Name: "r1"}, {Name: "r2"}},
				Type:  ast.NewIdent("any"),
			}},
		}})
	}
	return List
}

func isGinContext(expr ast.Expr) bool {
	t, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	t1, ok := t.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if t2, ok := t1.X.(*ast.Ident); !ok {
		return false
	} else if t1.Sel.Name != "Context" || t2.Name != "gin" {
		return false
	}
	return true
}

func (p *Parser) PrepareOneArgument(fn *ast.FuncDecl, arg *ast.Field, n int, List []ast.Stmt) []ast.Stmt {
	ni := ast.NewIdent(fmt.Sprintf("n%d", n))
	t, ok := arg.Type.(*ast.StarExpr)
	if !ok {
		fmt.Printf("Handler function params must be pointer, but %s has %s", fn.Name.Name, arg.Type)
		return List
	}
	if isGinContext(arg.Type) {
		List = append(List, &ast.AssignStmt{
			Lhs: []ast.Expr{ni},
			Rhs: []ast.Expr{ast.NewIdent("c")},
			Tok: token.DEFINE,
		})
	} else {
		List = append(List, &ast.AssignStmt{
			Lhs: []ast.Expr{ni},
			Rhs: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("new"), Args: []ast.Expr{t.X}}},
			Tok: token.DEFINE,
		})
		List = append(List, &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("e")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun:  &ast.SelectorExpr{Sel: &ast.Ident{Name: "BindJSON"}, X: &ast.Ident{Name: "c"}},
					Args: []ast.Expr{ni},
				}},
			},
			Cond: &ast.BinaryExpr{X: ast.NewIdent("e"), Op: token.NEQ, Y: ast.NewIdent("nil")},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				ginResponse(false, "", "e"),
				&ast.ReturnStmt{},
			}},
		})
	}
	return List
}

func (p *Parser) PrepareCall(fn *ast.FuncDecl, List []ast.Stmt) []ast.Stmt {
	var args []ast.Expr
	if fn.Type.Params.NumFields() == 1 {
		args = []ast.Expr{ast.NewIdent("n1")}
	} else if fn.Type.Params.NumFields() == 2 {
		args = []ast.Expr{
			ast.NewIdent("n1"),
			ast.NewIdent("n2"),
		}
	}
	if fn.Type.Results.NumFields() == 0 {
		List = append(List, &ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent(fn.Name.Name), Args: args}})
	} else if fn.Type.Results.NumFields() == 1 {
		List = append(List, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("r1")},
			Rhs: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent(fn.Name.Name), Args: args}},
			Tok: token.ASSIGN,
		})
	} else if fn.Type.Results.NumFields() == 2 {
		List = append(List, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("r1"),
				ast.NewIdent("r2"),
			},
			Rhs: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent(fn.Name.Name), Args: args}},
			Tok: token.ASSIGN,
		})
	}
	return List
}

func ginResponse(ok bool, data string, err string) *ast.ExprStmt {
	ok_v := "true"
	if !ok {
		ok_v = "false"
	}
	d := []ast.Expr{&ast.KeyValueExpr{
		Key:   &ast.BasicLit{Kind: token.STRING, Value: "\"ok\""},
		Value: ast.NewIdent(ok_v),
	}}
	if data != "" {
		d = append(d, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: "\"data\""},
			Value: ast.NewIdent(data),
		})
	}
	if err != "" {
		d = append(d, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: "\"error\""},
			Value: ast.NewIdent(err),
		})
	}
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{Sel: &ast.Ident{Name: "JSON"}, X: &ast.Ident{Name: "c"}},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: "200"},
				&ast.CompositeLit{
					Type: &ast.SelectorExpr{Sel: &ast.Ident{Name: "H"}, X: &ast.Ident{Name: "gin"}},
					Elts: d,
				},
			},
		},
	}
}

func ginResponseCheckError(data, err string) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent(err),
			Op: token.EQL,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{ginResponse(true, data, "")}},
		Else: &ast.BlockStmt{List: []ast.Stmt{ginResponse(false, "", err)}},
	}
}

func (p *Parser) PrepareReturns(fn *ast.FuncDecl, List []ast.Stmt) []ast.Stmt {
	if fn.Type.Results.NumFields() == 0 {
		List = append(List, ginResponse(true, "", ""))
	} else if fn.Type.Results.NumFields() == 1 {
		List = append(List, &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("_"), ast.NewIdent("ok")},
				Rhs: []ast.Expr{&ast.TypeAssertExpr{
					X:    ast.NewIdent("r1"),
					Type: ast.NewIdent("error"),
				}},
				Tok: token.DEFINE,
			},
			Cond: ast.NewIdent("ok"),
			Body: &ast.BlockStmt{List: []ast.Stmt{ginResponseCheckError("", "r1")}},
			Else: &ast.BlockStmt{List: []ast.Stmt{ginResponse(true, "r1", "")}},
		})
	} else {
		List = append(List, ginResponseCheckError("r1", "r2"))
	}
	return List
}

func (p *Parser) GenerateAddRouterFunc() {
	decl := &ast.FuncDecl{}
	decl.Name = &ast.Ident{Name: "AutoAPI_add_router_" + strings.TrimSuffix(p.Name, ".go")}
	decl.Type = &ast.FuncType{}
	decl.Type.Params = &ast.FieldList{
		List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("group")},
			Type:  &ast.SelectorExpr{Sel: &ast.Ident{Name: "RouterGroup"}, X: &ast.Ident{Name: "gin"}},
		}},
	}
	decl.Body = &ast.BlockStmt{List: []ast.Stmt{}}
	for _, f := range p.FuncInfoList {
		decl.Body.List = append(decl.Body.List, &ast.ExprStmt{X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{Sel: &ast.Ident{Name: "POST"}, X: &ast.Ident{Name: "group"}},
			Args: []ast.Expr{
				&ast.BasicLit{Value: "\"/autoapi/" + f.Name + "\"", Kind: token.STRING},
				ast.NewIdent("AutoAPI_handler_" + f.Name),
			},
		}})
	}
	p.HandlerFunc = append(p.HandlerFunc, decl)
}

func (p *Parser) GenerateJavaScript(js io.StringWriter) {
	template, err := os.ReadFile("template.js")
	if err != nil {
		fmt.Printf("Cannot open template.js: %s\n", err.Error())
		return
	}
	for _, f := range p.FuncInfoList {
		name := f.Name
		x := strings.ReplaceAll(string(template), "GetUser", name)
		_, _ = js.WriteString(x)
		_, _ = js.WriteString("\n")
	}
}
