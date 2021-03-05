package gogen

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/tools/goctl/api/spec"
	"github.com/tal-tech/go-zero/tools/goctl/config"
	ctlutil "github.com/tal-tech/go-zero/tools/goctl/util"
	"github.com/tal-tech/go-zero/tools/goctl/util/format"
	"github.com/tal-tech/go-zero/tools/goctl/vars"
)

const logicTemplate = `package logic

import (
	{{.ImportPkg}}
)

type {{.LogicContextName}} struct {
	logx.Logger
	ctx    context.Context
}

func New{{.LogicContextName}}(ctx context.Context) {{.LogicContextName}} {
	return {{.LogicContextName}}{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
	}
}

{{ range $method := .LogicMetas}} 

func (l *{{$.LogicContextName}}) {{.Function}}({{.Request}}) {{.ResponseType}} {
	// todo: add your logic here and delete this line

	{{.ReturnString}}
}
{{- end}}
`

func genLogic(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, group := range api.Service.Groups {
		name := fmt.Sprintf("%sLogicImpl_%s", strings.Title(group.GetAnnotation(groupProperty)),
			time.Now().Format("0102150405"))
		filename, err := format.FileNamingFormat(cfg.NamingFormat, name)
		if err != nil {
			return err
		}
		//合并生成一个Logic文件
		if err := genLogicByRoute(dir, filename, cfg, group, group.Routes); nil != err {
			return err
		}
	}
	return nil
}

type logicInfo struct {
	ImportPkg        string
	LogicContextName string
	LogicMetas       []logicMeta
}

type logicMeta struct {
	LogicMethodName string
	Function        string
	Request         string
	ResponseType    string
	ReturnString    string
}

func genLogicByRoute(dir, fileName string, cfg *config.Config, group spec.Group, route []spec.Route) error {
	//获取父级目录
	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	//获取子目录
	subPkg := getLogicFolderPath(group, route[0])

	//生成导入文件
	importPkg := genLogicImports(route, parentPkg)

	//逻辑上下文名称
	logicContextName := strings.Title(group.GetAnnotation(groupProperty)) + "LogicImpl"

	//逻辑层方法
	logicMetas := make([]logicMeta, 0)
	for _, r := range route {
		//逻辑层方法名称
		logicMethodName := getLogicName(r)

		var responseString string
		var returnString string
		var requestString string
		if len(r.ResponseTypeName()) > 0 {
			resp := responseGoTypeName(r, typesPacket)
			responseString = "(" + resp + ", error)"
			if strings.HasPrefix(resp, "*") {
				returnString = fmt.Sprintf("return &%s{}, nil", strings.TrimPrefix(resp, "*"))
			} else {
				returnString = fmt.Sprintf("return %s{}, nil", resp)
			}
		} else {
			responseString = "error"
			returnString = "return nil"
		}
		if len(r.RequestTypeName()) > 0 {
			requestString = "req " + requestGoTypeName(r, typesPacket)
		}

		logicMetas = append(logicMetas, logicMeta{
			LogicMethodName: logicMethodName,
			Function:        strings.Title(strings.TrimSuffix(logicMethodName, "Logic")),
			ResponseType:    responseString,
			ReturnString:    returnString,
			Request:         requestString,
		})
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          subPkg,
		filename:        fileName + ".go",
		templateName:    "logicTemplate",
		category:        category,
		templateFile:    logicTemplateFile,
		builtinTemplate: logicTemplate,
		data: logicInfo{
			LogicContextName: logicContextName,
			ImportPkg:        importPkg,
			LogicMetas:       logicMetas,
		},
	})
}

func getLogicFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return logicDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(logicDir, folder)
}

func genLogicImports(route []spec.Route, parentPkg string) string {
	var imports []string
	imports = append(imports, `"context"`+"\n")
	var isNeedImportTypePkg bool
	for _, r := range route {
		if len(r.ResponseTypeName()) > 0 || len(r.RequestTypeName()) > 0 {
			isNeedImportTypePkg = true
			break
		}
	}
	if isNeedImportTypePkg {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", ctlutil.JoinPackages(parentPkg, typesDir)))
	}
	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"", vars.ProjectOpenSourceURL))
	return strings.Join(imports, "\n\t")
}
