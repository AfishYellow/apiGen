package routers

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	ajson  = "application/json"
	axml   = "application/xml"
	aplain = "text/plain"
	ahtml  = "text/html"
	aform  = "multipart/form-data"
)

const (
	astTypeArray  = "array"
	astTypeObject = "object"
	astTypeMap    = "map"
)

var pkgCache map[string]struct{} //pkg:controller:function:comments comments: key:value
var controllerComments map[string]string
var importlist map[string]string
var astPkgs []*ast.Package
var globalBasePrefixUrlMap map[string]string
var globalPathName string
var handlers map[string]func(string, map[string]string)

func init() {
	pkgCache = make(map[string]struct{})
	controllerComments = make(map[string]string)
	importlist = make(map[string]string)
	astPkgs = make([]*ast.Package, 0)
	handlers = make(map[string]func(string, map[string]string))
}

// AddHandler  will add func to handler the comment with special prefix
func AddHandler(prefix string, handle func(item string, m map[string]string)) {
	handlers[prefix] = handle
}

// GenApiJson
func GenApiJson(curpath, destDirPath, pathName string) {
	//identify url pathName so that it can add the prefix url
	globalPathName = pathName
	routerPath := "/routers/router.go"

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, curpath+routerPath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Failed to open File :"+curpath+routerPath, err.Error())
	}

	maps := make([]map[string]string, 0)

	// Analyse controller package
	for _, im := range f.Imports {
		localName := ""
		if im.Name != nil {
			localName = im.Name.Name
		}
		m := analyseControllerPkg(localName, im.Path.Value, f)
		maps = append(maps, m...)

	}

	err = os.Mkdir(path.Join(curpath, destDirPath), 0666)
	if err != nil {
		fmt.Println("Failed to create Dir", err.Error())
	}
	fd, _ := os.Create(path.Join(curpath, destDirPath, "api.json"))
	dt, _ := json.Marshal(maps)
	_, err = fd.Write(dt)

}

//gen PrefixMap
func genGlobalPrefixMap(f *ast.File) {
	for _, d := range f.Decls {
		switch specDecl := d.(type) {
		case *ast.FuncDecl:
			for _, l := range specDecl.Body.List {
				switch stmt := l.(type) {
				case *ast.AssignStmt:
					for _, l := range stmt.Rhs {
						if v, ok := l.(*ast.CallExpr); ok {
							// Analyze NewNamespace, it will return version and the subfunction

							if getFuncName(v) != "NewNamespace" {
								continue
							}
							version, params := analyseNewNamespace(v)
							globalBasePrefixUrlMap = make(map[string]string, len(params))
							for _, p := range params {
								switch pp := p.(type) {
								case *ast.CallExpr:
									if selname := getFuncName(pp); selname == "NSNamespace" {
										s, params := analyseNewNamespace(pp)
										for _, sp := range params {
											switch pp := sp.(type) {
											case *ast.CallExpr:
												if getFuncName(pp) == "NSInclude" {
													globalBasePrefixUrlMap[analyseNSInclude(s, pp)] = version + s
												}
											}
										}
										// indicate the api was not in the folder
										// so the prefix url is /(version)
									} else if selname == "NSInclude" {
										globalBasePrefixUrlMap[analyseNSInclude("", pp)] = version
									}
								}
							}
						}

					}
				}
			}
		}
	}
}

func getFuncName(expr *ast.CallExpr) (name string) {
	switch sexpr := expr.Fun.(type) {
	case *ast.SelectorExpr:
		return sexpr.Sel.Name
	case *ast.Ident:
		return sexpr.Name
	}
	return ""

}

// analyseNewNamespace returns version and the others params
func analyseNewNamespace(ce *ast.CallExpr) (first string, others []ast.Expr) {
	for i, p := range ce.Args {
		if i == 0 {
			switch pp := p.(type) {
			case *ast.BasicLit:
				first = strings.Trim(pp.Value, `"`)
			}
			continue
		}
		others = append(others, p)
	}
	return
}

//
func analyseNSInclude(baseurl string, ce *ast.CallExpr) string {
	cname := ""
	for _, p := range ce.Args {
		var x *ast.SelectorExpr
		var p1 interface{} = p
		if ident, ok := p1.(*ast.Ident); ok {
			if assign, ok := ident.Obj.Decl.(*ast.AssignStmt); ok {
				if len(assign.Rhs) > 0 {
					p1 = assign.Rhs[0].(*ast.UnaryExpr)
				}
			}
		}
		if _, ok := p1.(*ast.UnaryExpr); ok {
			x = p1.(*ast.UnaryExpr).X.(*ast.CompositeLit).Type.(*ast.SelectorExpr)
		} else {
			//beeLogger.Log.Warnf("Couldn't determine type\n")
			continue
		}
		if v, ok := importlist[fmt.Sprint(x.X)]; ok {
			cname = v + x.Sel.Name
		}
	}
	return cname
}

//
func analyseControllerPkg(localName, pkgpath string, f *ast.File) []map[string]string {

	pkgpath = strings.Trim(pkgpath, "\"")
	if isSystemPackage(pkgpath) {
		return nil
	}
	if pkgpath == "github.com/astaxie/beego" {
		return nil
	}
	if localName != "" {
		importlist[localName] = pkgpath
	} else {
		pps := strings.Split(pkgpath, "/")
		importlist[pps[len(pps)-1]] = pkgpath
	}

	genGlobalPrefixMap(f)

	pkg, err := build.Default.Import(pkgpath, ".", build.FindOnly)
	if err != nil {
		//beeLogger.Log.Fatalf("Package %s cannot be imported: %v", pkgpath, err)
	}
	pkgRealpath := pkg.Dir
	if pkgRealpath != "" {
		if _, ok := pkgCache[pkgpath]; ok {
			return nil
		}
		pkgCache[pkgpath] = struct{}{}
	} else {
		//beeLogger.Log.Fatalf("Package '%s' does not have source directory", pkgpath)
	}

	fileSet := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fileSet, pkgRealpath, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
	}, parser.ParseComments)
	if err != nil {
		//beeLogger.Log.Fatalf("Error while parsing dir at '%s': %s", pkgpath, err)
	}

	m := make([]map[string]string, 0)

	for _, pkg := range astPkgs {
		for _, fl := range pkg.Files {
			for _, d := range fl.Decls {
				switch specDecl := d.(type) {
				case *ast.FuncDecl:
					if specDecl.Recv != nil && len(specDecl.Recv.List) > 0 {
						if t, ok := specDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
							// Parse controller method
							comments, err := parserComments(specDecl, fmt.Sprint(t.X), pkgpath)
							if err != nil {
							}
							m = append(m, comments)

						}
					}
				//	gen the comments of Controller
				case *ast.GenDecl:
					if specDecl.Tok == token.TYPE {
						for _, s := range specDecl.Specs {
							switch tp := s.(*ast.TypeSpec).Type.(type) {
							case *ast.StructType:
								_ = tp.Struct
								// Parse controller definition comments
								if strings.TrimSpace(specDecl.Doc.Text()) != "" {
									controllerComments[pkgpath+s.(*ast.TypeSpec).Name.String()] = specDecl.Doc.Text()
								}
							}
						}
					}
				}
			}
		}
	}
	return m
}

//
func isSystemPackage(pkgpath string) bool {
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		goroot = runtime.GOROOT()
	}
	if goroot == "" {
		//beeLogger.Log.Fatalf("GOROOT environment variable is not set or empty")
	}

	wg, _ := filepath.EvalSymlinks(filepath.Join(goroot, "src", "pkg", pkgpath))
	if FileExists(wg) {
		return true
	}

	//TODO(zh):support go1.4
	wg, _ = filepath.EvalSymlinks(filepath.Join(goroot, "src", pkgpath))
	return FileExists(wg)
}

// parse the func comments
func parserComments(f *ast.FuncDecl, controllerName, pkgpath string) (m map[string]string, err error) {
	var routerPath string

	comments := f.Doc

	m = make(map[string]string)

	//TODO: resultMap := buildParamMap(f.Type.Results)
	if comments != nil && comments.List != nil {
		for _, c := range comments.List {

			s := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if strings.HasPrefix(s, "@") {
				s = strings.TrimSpace(strings.TrimPrefix(s, "@"))

				s = strings.ReplaceAll(s, "\t", " ")

				firstIndex := strings.Index(s, " ")

				if firstIndex == -1 {
					fmt.Println("can not find space")
				} else {

					if f := handlers[s[:firstIndex]]; f != nil {
						f(s[firstIndex+1:], m)
					} else {
						m[s[:firstIndex]] = strings.TrimSpace(s[firstIndex+1:])
					}
					if strings.HasPrefix(s, globalPathName) {
						if m[globalPathName] == "" {
							routerPath = strings.TrimSpace(s[firstIndex+1:])
						} else {
							routerPath = m[globalPathName]
						}
					}
				}

			}
		}
	}
	routerPath = urlReplace(routerPath)

	m[globalPathName] = globalBasePrefixUrlMap[pkgpath+controllerName] + m[globalPathName]

	return m, nil
}

//replace :id to {id}
func urlReplace(src string) string {
	pt := strings.Split(src, "/")
	for i, p := range pt {
		if len(p) > 0 {
			if p[0] == ':' {
				pt[i] = "{" + p[1:] + "}"
			} else if p[0] == '?' && p[1] == ':' {
				pt[i] = "{" + p[2:] + "}"
			}

			if pt[i][0] == '{' && strings.Contains(pt[i], ":") {
				pt[i] = pt[i][:strings.Index(pt[i], ":")] + "}"
			} else if pt[i][0] == '{' && strings.Contains(pt[i], "(") {
				pt[i] = pt[i][:strings.Index(pt[i], "(")] + "}"
			}
		}
	}
	return strings.Join(pt, "/")
}

// LinkNamespace used as link action
type LinkNamespace func(*Namespace)

type Namespace struct{}

func NewNamespace(prefix string, params ...LinkNamespace) *Namespace {
	return &Namespace{}
}

// NSNamespace add sub Namespace
func NSNamespace(prefix string, params ...LinkNamespace) LinkNamespace {
	return func(ns *Namespace) {
	}
}

func NSInclude(cList ...interface{}) LinkNamespace {
	return func(ns *Namespace) {
	}
}

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
