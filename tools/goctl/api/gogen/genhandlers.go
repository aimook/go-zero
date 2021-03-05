package gogen

import (
	"fmt"
	"github.com/tal-tech/go-zero/tools/goctl/util/format"
	"path"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/tools/goctl/api/spec"
	"github.com/tal-tech/go-zero/tools/goctl/config"
	"github.com/tal-tech/go-zero/tools/goctl/util"
	"github.com/tal-tech/go-zero/tools/goctl/vars"
)

const handlerTemplate = `package handler

import (
	"net/http"

	{{.ImportPackages}}
)
{{range $handler:= .HandlerMetas}}

//{{.HandlerName}} {{.Summary}}
func {{.HandlerName}}() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Create(w).Error(err)
			return
		}{{end}}

		l := logic.New{{$.LogicContextName}}(r.Context())
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}req{{end}})
		if err != nil {
			httpx.Create(w).Error(err)
		} else {
			{{if .HasResp}}httpx.Create(w).Data(resp).Success(){{else}}httpx.Create(w).Success(){{end}}
		}
	}
}

{{- end}}
`

type handlerInfo struct {
	LogicContextName string
	ImportPackages   string
	HandlerMetas     []handlerMeta
}

type handlerMeta struct {
	Summary     string
	HandlerName string
	RequestType string
	Call        string
	HasResp     bool
	HasRequest  bool
}

func genHandler(dir, handlerFileName string, cfg *config.Config, group spec.Group, route []spec.Route) error {
	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}
	handlerMetas := make([]handlerMeta, 0)
	importPackages := genHandlerImports(group, route, parentPkg)

	//逻辑上下文名称
	logicContextName := strings.Title(group.GetAnnotation(groupProperty)) + "LogicImpl"

	for _, r := range route {
		meta := handlerMeta{}
		//Handler名称
		handler := getHandlerName(r)
		if getHandlerFolderPath(group, r) != handlerDir {
			handler = strings.Title(handler)
		}
		meta.HandlerName = handler

		//Handler注释
		summary, ok := r.AtDoc.Properties["summary"]
		if ok {
			summary = strings.TrimPrefix(summary, "\"")
			summary = strings.TrimSuffix(summary, "\"")
			meta.Summary = summary
		}
		//请求类型
		meta.RequestType = util.Title(r.RequestTypeName())
		//调用方法名称
		meta.Call = strings.Title(strings.TrimSuffix(handler, "Handler"))
		//是否有返回值
		meta.HasResp = len(r.ResponseTypeName()) > 0
		//是否有请求参数
		meta.HasRequest = len(r.RequestTypeName()) > 0

		handlerMetas = append(handlerMetas, meta)
	}

	return doGenToFile(dir, handlerFileName, group, route[0], handlerInfo{
		LogicContextName: logicContextName,
		ImportPackages:   importPackages,
		HandlerMetas:     handlerMetas,
	})
}

func doGenToFile(dir, fileName string, group spec.Group, route spec.Route, handleObj handlerInfo) error {
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          getHandlerFolderPath(group, route),
		filename:        fileName + ".go",
		templateName:    "handlerTemplate",
		category:        category,
		templateFile:    handlerTemplateFile,
		builtinTemplate: handlerTemplate,
		data:            handleObj,
	})
}

func genHandlers(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, group := range api.Service.Groups {
		name := fmt.Sprintf("%sHandler_%s", strings.Title(group.GetAnnotation(groupProperty)),
			time.Now().Format("0102150405"))
		filename, err := format.FileNamingFormat(cfg.NamingFormat, name)
		if err != nil {
			return err
		}
		//合并生成一个Handler文件
		if err := genHandler(dir, filename, cfg, group, group.Routes); err != nil {
			return err
		}
	}
	return nil
}

func genHandlerImports(group spec.Group, route []spec.Route, parentPkg string) string {
	var imports []string
	//导入逻辑层包(每组导入1次即可)
	imports = append(imports, fmt.Sprintf("\"%s\"",
		util.JoinPackages(parentPkg, getLogicFolderPath(group, route[0]))))
	//标记是否需要导入type包
	var isNeedImportRequestTypePkg bool
	for _, r := range route {
		//只要有一个handler引用type包，则应导入
		if len(r.RequestTypeName()) > 0 {
			isNeedImportRequestTypePkg = true
			break
		}
	}
	if isNeedImportRequestTypePkg {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", util.JoinPackages(parentPkg, typesDir)))
	}
	imports = append(imports, fmt.Sprintf("\"%s/rest/httpx\"", vars.ProjectOpenSourceURL))

	return strings.Join(imports, "\n\t")
}

func getHandlerBaseName(route spec.Route) (string, error) {
	handler := route.Handler
	handler = strings.TrimSpace(handler)
	handler = strings.TrimSuffix(handler, "handler")
	handler = strings.TrimSuffix(handler, "Handler")
	return handler, nil
}

//getHandlerFolderPath
func getHandlerFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return handlerDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(handlerDir, folder)
}

func getHandlerName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Handler"
}

func getLogicName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Logic"
}
