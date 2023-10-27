package terminal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eviltomorrow/toolbox/apps/minshell/assets"
	"github.com/eviltomorrow/toolbox/lib/system"
	"github.com/olekukonko/tablewriter"
)

type Option struct {
	ShowFooter    bool
	FooterContent string
}

func RenderTableFromFile(path string) error {
	machineFile := path
	if machineFile == "" {
		machineFile = filepath.Join(system.Runtime.RootDir, "etc", "machines.xlsx")
	}
	machines, err := assets.LoadFile(machineFile)
	if err != nil {
		return err
	}
	RenderTable(machines, Option{ShowFooter: true})
	return nil
}

func RenderTable(machines []*assets.Machine, option Option) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"No", "IP", "NAT-IP", "Port", "User", "Password", "PrivateKey-Path"})

	data := [][]string{}
	if len(machines) == 0 {
		data = append(data, []string{"Null", "Null", "Null", "Null", "Null", "Null", "Null"})
	} else {
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
			line := make([]string, 0, 7)
			line = append(line, fmt.Sprintf("%3d", i+1))
			line = append(line, machine.IP)
			line = append(line, machine.NatIP)
			line = append(line, fmt.Sprintf("%d", machine.Port))
			line = append(line, machine.Username)
			line = append(line, password)
			line = append(line, privateKeyPath)
			data = append(data, line)
		}
	}

	if option.ShowFooter {
		table.SetFooter([]string{"", "", "", "", "", "Total", fmt.Sprintf("%3d", len(machines))})
		table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)
	}

	table.SetBorder(true)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	if option.FooterContent != "" {
		fmt.Println(option.FooterContent)
	}
	fmt.Println()
}
