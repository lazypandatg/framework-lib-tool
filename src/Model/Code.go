package Model

import (
	"bytes"
	"fmt"
	"github.com/fatih/structtag"
	"go/ast"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Code struct {
	Path          []string
	AccessTagName string
	TypeList      []string
}

func (_this *Code) ParsePackage() []*packages.Package {
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	}
	pkgList, err := packages.Load(cfg, _this.Path...)
	if err != nil {
		log.Fatal(err)
	}
	return pkgList
}

func (_this *Code) GeneratePackage() {
	packageList := _this.ParsePackage()
	for i := 0; i < len(packageList); i++ {
		_this.Generate(packageList[i])
	}
}
func (_this *Code) TransformName(name string) string {
	mark := strings.LastIndex(name, "Model")
	if mark == -1 || mark+5 != len(name) {
		return name
	}
	return name[:mark] + "Core"
}
func (_this *Code) FindProjectName(path string) string {
	for true {
		c, err := ioutil.ReadFile(filepath.Dir(path) + "/go.mod")
		if err == nil {
			str := string(c)
			str = str[strings.Index(str, "module")+7 : strings.Index(str, "\n")]
			return str
		} else {
			path = filepath.Dir(path)
		}
	}
	return ""
}
func (_this *Code) Generate(pack *packages.Package) {
	if len(pack.GoFiles) == 0 {
		return
	}
	path := filepath.Dir(filepath.Dir(pack.GoFiles[0])) + "/"
	_this.FindProjectName(path)
	_, err := os.Stat(path + "Core")
	if err != nil {
		err = os.MkdirAll(path+"Core", 0777)
		if err != nil {
			return
		}
	}
	ProjectName := _this.FindProjectName(pack.GoFiles[0])
	packName := _this.TransformName(pack.Name)
	if pack.Name == packName {
		return
	}
	for _, file := range pack.Syntax {
		structMap := _this.ParseStruct(file, pack.Fset, _this.AccessTagName)
		for stName, info := range structMap {
			codeWrite := CodeWrite{}
			codeWrite.Println("// Package WebSiteCore Auto /**\n\n")
			codeWrite.Println("package %s\n\n", packName)
			codeWrite.Println(`import "` + ProjectName + `/src/Application/WebSite/Model"` + "\n")
			codeWrite.Println(`import "` + ProjectName + `/src/CoreManage"` + "\n")
			codeWrite.Println(`import "` + ProjectName + `/src/Lib/DataSource"` + "\n")
			codeWrite.Println(`import "` + ProjectName + `/src/Lib/Base"` + "\n\n")
			codeWrite.Println(`var %s = _%s { `+"\n", stName, stName)
			for _, field := range info {
				codeWrite.Println(`	%s: BaseLib.Param{Name: "%s"},`+"\n", field.Name, field.Name)
			}
			codeWrite.Println("}\n\n")
			codeWrite.Println("type _%s struct {\n", stName)
			for _, field := range info {
				codeWrite.Println("	%s\n", field.Name+" BaseLib.Param ")
			}
			codeWrite.Println("}\n\n")
			codeWrite.Println(`func (_this *_` + stName + `) Insert(list []DataSourceLib.FieldModel) (int64, error) {
	list = append(list, CoreManage.DataBase.Table("` + stName + `", ""))
	insert, err := CoreManage.DataBase.Insert(list)
	return insert, err
}` + "\n\n")
			codeWrite.Println(`func (_this *_` + stName + `) Update(list []DataSourceLib.FieldModel) (int64, error) {
	list = append(list, CoreManage.DataBase.Table("` + stName + `", ""))
	update, err := CoreManage.DataBase.Update(list)
	return update, err
}` + "\n\n")
			codeWrite.Println(`func (_this *_` + stName + `) Search(list []DataSourceLib.FieldModel) ([]WebSiteModel.Server, error) {
	list = append(list, CoreManage.DataBase.Table("` + stName + `", ""))
	var result []WebSiteModel.Server
	err := CoreManage.DataBase.Select(list, result, WebSiteModel.Server{})
	return result, err
}` + "\n\n")
			codeWrite.Println(`func (_this *_` + stName + `) Count(list []DataSourceLib.FieldModel) (int, error) {
	list = append(list, CoreManage.DataBase.Table("` + stName + `", ""))
	return CoreManage.DataBase.Count(list)
}` + "\n\n")
			log.Println(stName)
			//log.Println(path + "Field/" + _this.TransformName(file.Name.String()) + ".go")
			codeWrite.Write(path + "/Core/" + packName + stName + "Core.go")
		}
	}
}

func (_this *Code) ParseStruct(file *ast.File, token *token.FileSet, accessTagName string) map[string][]CodeStruct {
	structMap := make(map[string][]CodeStruct)
	collectStructs := func(x ast.Node) bool {
		ts, ok := x.(*ast.TypeSpec)
		if !ok || ts.Type == nil {
			return true
		}

		// 获取结构体名称
		structName := ts.Name.Name

		log.Println("类名:", structName)

		s, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		fileInfos := make([]CodeStruct, 0)
		for _, field := range s.Fields.List {
			if len(field.Names) == 0 {
				continue
			}
			name := field.Names[0].Name
			info := CodeStruct{Name: name}
			var typeNameBuf bytes.Buffer
			err := printer.Fprint(&typeNameBuf, token, field.Type)
			if err != nil {
				fmt.Println("获取类型失败:", err)
				continue
			}
			info.Type = typeNameBuf.String()
			if field.Tag != nil {
				tagList, err := structtag.Parse(strings.Trim(field.Tag.Value, "`"))
				if err != nil {
					continue
				}
				tag, err := tagList.Get(accessTagName)
				if err != nil {
					continue
				}
				info.Access = tag.String()
			}
			fileInfos = append(fileInfos, info)
		}
		structMap[structName] = fileInfos
		return false
	}
	ast.Inspect(file, collectStructs)
	return structMap
}
