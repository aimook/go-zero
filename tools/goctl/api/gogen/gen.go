package gogen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/tal-tech/go-zero/core/logx"
	apiformat "github.com/tal-tech/go-zero/tools/goctl/api/format"
	"github.com/tal-tech/go-zero/tools/goctl/api/parser"
	apiutil "github.com/tal-tech/go-zero/tools/goctl/api/util"
	"github.com/tal-tech/go-zero/tools/goctl/config"
	"github.com/tal-tech/go-zero/tools/goctl/util"
	"github.com/urfave/cli"
)

const tmpFile = "%s-%d"

var tmpDir = path.Join(os.TempDir(), "goctl")

// GoCommand gen go project files from command line
func GoCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
	namingStyle := c.String("style")

	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	return DoGenProject(apiFile, dir, namingStyle)
}

// DoGenProject gen go project files with api file
func DoGenProject(apiFile, dir, style string) error {
	api, err := parser.Parse(apiFile)
	if err != nil {
		return err
	}

	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}
	// 创建API模块目录
	logx.Must(util.MkdirIfNotExist(dir))
	// 生成etc配置文件
	logx.Must(genEtc(dir, cfg, api))
	// 生成配置文件
	logx.Must(genConfig(dir, cfg, api))
	// 生成入口文件
	logx.Must(genMain(dir, cfg, api))
	// 生成应用上下文
	logx.Must(genServiceContext(dir, cfg, api))
	// 生成type(模型)
	logx.Must(genTypes(dir, cfg, api))
	// 生成路由配置文件
	logx.Must(genRoutes(dir, cfg, api))
	// 生成Handler文件
	logx.Must(genHandlers(dir, cfg, api))
	// 生成逻辑层代码
	logx.Must(genLogic(dir, cfg, api))
	// 生成中间件(如果配置)
	logx.Must(genMiddleware(dir, cfg, api))

	if err := backupAndSweep(apiFile); err != nil {
		return err
	}

	if err := apiformat.ApiFormatByPath(apiFile); err != nil {
		return err
	}

	fmt.Println(aurora.Green("Done."))
	return nil
}

func backupAndSweep(apiFile string) error {
	var err error
	var wg sync.WaitGroup

	wg.Add(2)
	_ = os.MkdirAll(tmpDir, os.ModePerm)

	go func() {
		_, fileName := filepath.Split(apiFile)
		_, e := apiutil.Copy(apiFile, fmt.Sprintf(path.Join(tmpDir, tmpFile), fileName, time.Now().Unix()))
		if e != nil {
			err = e
		}
		wg.Done()
	}()
	go func() {
		if e := sweep(); e != nil {
			err = e
		}
		wg.Done()
	}()
	wg.Wait()

	return err
}

func sweep() error {
	keepTime := time.Now().AddDate(0, 0, -7)
	return filepath.Walk(tmpDir, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		pos := strings.LastIndexByte(info.Name(), '-')
		if pos > 0 {
			timestamp := info.Name()[pos+1:]
			seconds, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				// print error and ignore
				fmt.Println(aurora.Red(fmt.Sprintf("sweep ignored file: %s", fpath)))
				return nil
			}

			tm := time.Unix(seconds, 0)
			if tm.Before(keepTime) {
				if err := os.Remove(fpath); err != nil {
					fmt.Println(aurora.Red(fmt.Sprintf("failed to remove file: %s", fpath)))
					return err
				}
			}
		}

		return nil
	})
}
