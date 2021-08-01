package main

import (
	"CDNDrive/drivers"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	//初始化
	loadDrivers()
	colorLogger = &colorLogger_t{
		prefix: func() string { return "[" + FormatTime(time.Now().Unix()) + "]" },
	}

	flag_drivers := &cli.StringFlag{
		Name:    "driver",
		Aliases: []string{"d"},
		Usage: "上传种类，同时上传多个地方请用逗号分割，必须为以下的一个或多个： " + func() (txt string) {
			for i := 0; i < len(_drivers); i++ {
				txt += _drivers[i].Name()
				if i != len(_drivers)-1 {
					txt += ", "
				}
			}
			return
		}(),
	}
	flag_threadN := &cli.IntFlag{
		Name:    "thread",
		Aliases: []string{"t"},
		Usage:   "并发连接数",
		Value:   4,
	}

	app := &cli.App{
		Name:    "CDNDrive-go",
		Usage:   "Make Picbeds Great Cloud Storages!",
		Version: "v0.1",
		Authors: []*cli.Author{
			&cli.Author{
				Name: "猫村あおい",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "调试",
			}},
		Before: func(c *cli.Context) error {
			_debug = c.Bool("debug")
			return nil
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:    "download",
				Aliases: []string{"d"},
				Usage:   "下载文件",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "https",
						Usage: "强制使用https",
					}, &cli.BoolFlag{
						Name:  "batch",
						Usage: "批量下载模式",
					},
					flag_threadN,
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 && !c.Bool("batch") {
						cli.ShowCommandHelpAndExit(c, "download", 1)
					}
					HandlerDownload(c.Args().Slice(), c.Bool("https"), c.Int("thread"), c.Bool("batch"))
					return nil
				},
			}, &cli.Command{
				Name:    "upload",
				Aliases: []string{"u"},
				Usage:   "上传文件",
				Flags: []cli.Flag{
					flag_threadN,
					flag_drivers,
					&cli.IntFlag{
						Name:    "block-size",
						Aliases: []string{"b"},
						Usage:   "分块大小，单位为字节",
						Value:   4 * 1024 * 1024,
					},
				},
				Action: func(c *cli.Context) error {
					//driver过滤，顺便去重
					ds := make(map[string]drivers.Driver)
					for _, name := range strings.Split(c.String("driver"), ",") {
						_d := queryDriverByName(name)
						if _d != nil {
							ds[name] = _d
						}
					}

					if c.NArg() == 0 || len(ds) == 0 {
						cli.ShowCommandHelpAndExit(c, "upload", 1)
					}

					if _debug { //内存分析
						go func() {
							http.ListenAndServe("127.0.0.1:9459", nil)
						}()
					}

					HandlerUpload(c.Args().Slice(), ds, c.Int("thread"), c.Int("block-size"))
					return nil
				},
			}, &cli.Command{
				Name:    "cookie",
				Aliases: []string{"c"},
				Usage:   "有些图床上传需要登录，请提供小饼干。",
				Flags: []cli.Flag{
					flag_drivers,
					&cli.BoolFlag{
						Name:  "force",
						Usage: "强制设置，跳过 cookie 有效性检查",
					},
				},
				Action: func(c *cli.Context) error {
					//driver过滤，顺便去重
					ds := make(map[string]drivers.Driver)
					for _, name := range strings.Split(c.String("driver"), ",") {
						_d := queryDriverByName(name)
						if _d != nil {
							ds[name] = _d
						}
					}

					if c.NArg() == 0 || len(ds) == 0 {
						cli.ShowCommandHelpAndExit(c, "cookie", 1)
					}

					if len(ds) > 1 {
						fmt.Println("设置 cookie 时一次只能输入一个 driver")
						return nil
					}

					cookieJson := loadUserCookie()
					for name, _ := range ds {
						err := cookieJson.setDriveCookie(name, c.Args().Get(0), c.Bool("force"))
						if err == nil {
							fmt.Println("设置成功")
						} else {
							fmt.Println("设置失败", err.Error())
						}
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}