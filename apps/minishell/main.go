package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eviltomorrow/toolbox/apps/minishell/adapter"
	"github.com/eviltomorrow/toolbox/apps/minishell/assets"
	"github.com/eviltomorrow/toolbox/apps/minishell/terminal"
	"github.com/eviltomorrow/toolbox/lib/buildinfo"
	"github.com/eviltomorrow/toolbox/lib/system"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var (
	AppName     = "unknown"
	MainVersion = "unknown"
	GitSha      = "unknown"
	BuildTime   = "unknown"
)

func init() {
	buildinfo.AppName = AppName
	buildinfo.MainVersion = MainVersion
	buildinfo.GitSha = GitSha
	buildinfo.BuildTime = BuildTime
}

var (
	greenbold = color.New(color.FgGreen, color.Bold)
	red       = color.New(color.FgRed)
	redbold   = color.New(color.BgRed, color.Bold)
)

func main() {
	if err := system.LoadRuntime(); err != nil {
		log.Fatalf("LoadRuntime failure, nest error: %v", err)
	}

	app := &cli.App{
		Name:    buildinfo.AppName,
		Version: buildinfo.MainVersion,
		Authors: []*cli.Author{
			{
				Name:  "liarsa.localdomain",
				Email: "eviltomorrow@gmail.com",
			},
		},
		Copyright: "(c) 2023~2023 By Liarsa, All rights reserved.",
		Usage:     "快速登录 ssh server 工具",

		Commands: []*cli.Command{
			{
				Name:      "show",
				Usage:     "显示所有机器列表",
				UsageText: "./minishell show",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "the machines file path"},
					&cli.BoolFlag{Name: "print", Aliases: []string{"p"}, Usage: "show password"},
				},
				Action: func(cCtx *cli.Context) error {
					path := cCtx.String("file")
					print := cCtx.Bool("print")
					return terminal.RenderTableFromFile(path, print)
				},
			},

			{
				Name:      "version",
				Usage:     "打印版本信息",
				UsageText: "版本信息",
				Action: func(cCtx *cli.Context) error {
					fmt.Println(buildinfo.Version())
					return nil
				},
			},
		},
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Action: func(cCtx *cli.Context) error {
			args := cCtx.Args().Len()
			switch args {
			case 0:
				path := cCtx.String("file")
				return terminal.RenderTableFromFile(path, false)

			default:
				path := cCtx.String("file")
				machines, err := assets.LoadFile(path)
				if err != nil {
					return err
				}
				cond := cCtx.Args().First()
				machines, err = machines.Find(cond)
				if err == assets.ErrNotFound {
					redbold.Println("==> Error: 未找到指定 machine")
					return nil
				}
				if err != nil {
					redbold.Printf("==> Error: 查找主机失败, nest error: %v\r\n", err)
					return nil
				}
				if len(machines) == 1 {
					machine := machines[0]
					greenbold.Printf("==> Prepare to login [%s/%s]\r\n", machine.NatIP, machine.IP)
					fmt.Println()

					ip := machine.IP
					if machine.NatIP != "" && machine.NatIP != "无" {
						ip = machine.NatIP
					}
					var privateKeyPath string
					if machine.PrivateKeyPath != "" && machine.PrivateKeyPath != assets.NotExist {
						privateKeyPath = machine.PrivateKeyPath
					}
					var password string
					if machine.Password != "" && machine.Password != assets.NotExist {
						password = machine.Password
					}

					if err := adapter.InteractiveWithTerminalForSSH(machine.Username, password, privateKeyPath, ip, machine.Port, 10*time.Second, strings.EqualFold(machine.Device, "linux")); err != nil {
						greenbold.Printf("==> Fatal: Login resource failure, nest error: %v, resource: %v\r\n", err, ip)
						fmt.Println()
						os.Exit(1)
					}
					greenbold.Println("==> Logout")
					return nil
				}

				machinesWrapper := make([]*assets.Machine, 0, len(machines))
				for _, machine := range machines {
					machinesWrapper = append(machinesWrapper, &assets.Machine{
						NatIP:          strings.ReplaceAll(machine.NatIP, cond, red.Sprintf("%s", cond)),
						IP:             strings.ReplaceAll(machine.IP, cond, red.Sprintf("%s", cond)),
						Username:       machine.Username,
						Password:       machine.Password,
						Port:           machine.Port,
						Timeout:        machine.Timeout,
						PrivateKeyPath: machine.PrivateKeyPath,
					})
				}
				terminal.RenderTable(machinesWrapper, terminal.Option{FooterContent: greenbold.Sprintf("==> Warn: 包含多台 machine, 请指定一台 machine")})
				return nil
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
