package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eviltomorrow/toolbox/app/minshell/adapter"
	"github.com/eviltomorrow/toolbox/app/minshell/assets"
	"github.com/eviltomorrow/toolbox/lib/buildinfo"
	"github.com/eviltomorrow/toolbox/lib/system"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
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

func main() {
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
				UsageText: "./minshell show",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "the machines.xlsx path"},
				},
				Action: func(cCtx *cli.Context) error {
					var path = cCtx.String("file")
					return renderTable(path)
				},
			},

			{
				Name:      "version",
				Usage:     "打印版本信息",
				UsageText: "版本信息",
				Action: func(cCtx *cli.Context) error {
					fmt.Println(buildinfo.GetVersion())
					return nil
				},
			},
		},
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Action: func(cCtx *cli.Context) error {
			if cCtx.Args().Len() == 0 {
				var path = cCtx.String("file")
				return renderTable(path)
			} else {
				var machineFile = filepath.Join(system.Runtime.RootDir, "etc", "machines.xlsx")
				machines, err := assets.LoadFile(machineFile)
				if err != nil {
					return err
				}

				var cond = cCtx.Args().First()
				machine, err := machines.Find(cond)
				if err == assets.ErrNotFound {
					fmt.Println("Error: 未找到指定 machine")
					return nil
				}
				if err != nil {
					fmt.Printf("Error: 查找主机失败, nest error: %v, input: %v", err, cond)
					return nil
				}

				greenbold.Printf("==> Prepare to login [%s/%s]\r\n", machine.NatIP, machine.IP)
				fmt.Println()

				var ip = machine.IP
				if machine.NatIP != "" && machine.NatIP != "无" {
					ip = machine.NatIP
				}
				var privateKeyPath string
				if machine.PrivateKeyPath != "" && machine.PrivateKeyPath != "无" {
					privateKeyPath = machine.PrivateKeyPath
				}

				if err := adapter.InteractiveWithTerminalForSSH(machine.Username, machine.Password, privateKeyPath, ip, machine.Port, 10*time.Second); err != nil {
					log.Fatalf("Login resource failure, nest error: %v, resource: %v", err, ip)
				}
				greenbold.Println("==> Logout")
				return nil
			}

		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

var (
	greenbold = color.New(color.FgGreen, color.Bold)
)

func renderTable(path string) error {
	var machineFile = path
	if machineFile == "" {
		machineFile = filepath.Join(system.Runtime.RootDir, "etc", "machines.xlsx")
	}
	machines, err := assets.LoadFile(machineFile)
	if err != nil {
		return err
	}

	var data = [][]string{}
	for i, machine := range machines {
		var (
			password       = "********"
			privateKeyPath = "********"
		)
		if machine.Password == "" || machine.Password == assets.NotExist {
			password = machine.Password
		}
		if machine.PrivateKeyPath == "" || machine.PrivateKeyPath == assets.NotExist {
			privateKeyPath = machine.PrivateKeyPath
		}
		var line = make([]string, 0, 7)
		line = append(line, fmt.Sprintf("%3d", i+1))
		line = append(line, machine.IP)
		line = append(line, machine.NatIP)
		line = append(line, fmt.Sprintf("%d", machine.Port))
		line = append(line, machine.Username)
		line = append(line, password)
		line = append(line, privateKeyPath)
		data = append(data, line)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"No", "IP", "NAT-IP", "Port", "User", "Password", "PrivateKey-Path"})
	table.SetFooter([]string{"", "", "", "", "", "Total", fmt.Sprintf("%3d", len(machines))})
	table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)
	table.SetBorder(true)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
	return nil
}
