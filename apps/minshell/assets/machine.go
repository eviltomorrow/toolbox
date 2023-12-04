package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/eviltomorrow/toolbox/lib/system"
	"github.com/xuri/excelize/v2"
)

var ErrNotFound = errors.New("not found")

const (
	NotExist = "无"
)

type MachineList []*Machine

type Machine struct {
	NatIP          string        `toml:"nat-ip" json:"nat-ip"`
	IP             string        `toml:"ip" json:"ip"`
	Username       string        `toml:"username" json:"username"`
	Password       string        `toml:"password" json:"password"`
	Port           int           `toml:"port" json:"port"`
	Timeout        time.Duration `json:"timeout"`
	PrivateKeyPath string        `toml:"private-key" json:"private-key"`
	Device         string        `toml:"device" json:"device"`
}

func (m *Machine) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

func LoadFile(path string) (MachineList, error) {
	machineFile := strings.TrimSpace(path)

	if machineFile == "" {
		entries, err := os.ReadDir(filepath.Join(system.Runtime.RootDir, "etc"))
		if err != nil {
			return nil, err
		}

		fs := make([]string, 0, len(entries))
		for _, entry := range entries {
			name := entry.Name()
			fs = append(fs, filepath.Join(system.Runtime.RootDir, "etc", name))
		}

		if len(fs) == 0 {
			return nil, fmt.Errorf("no valid machines file")
		}
		machineFile = fs[0]
	}
	if strings.HasSuffix(machineFile, ".xlsx") {
		return loadExcelFile(machineFile)
	}
	if strings.HasSuffix(machineFile, ".conf") {
		return LoadTomlFile(machineFile)
	}
	return nil, fmt.Errorf("not support file, path: %v", path)
}

func LoadTomlFile(path string) ([]*Machine, error) {
	type F struct {
		Machines []*Machine `toml:"machines" json:"machines"`
	}

	f := new(F)
	if _, err := toml.DecodeFile(path, f); err != nil {
		return nil, err
	}
	return f.Machines, nil
}

func loadExcelFile(path string) ([]*Machine, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	var (
		rowCount, colCount int
		line               = make([]string, 0, 16)
		machines           = make([]*Machine, 0, 128)
	)
loop:
	for _, row := range rows {
		rowCount++
		if len(row) > 0 && row[0] == "" {
			continue
		}
		for _, col := range row {
			col = strings.TrimSpace(col)
			if rowCount == 1 || rowCount == 2 {
				continue loop
			}
			if colCount >= 7 {
				break
			}

			line = append(line, col)
			colCount++
		}
		colCount = 0
		port, err := strconv.ParseInt(line[2], 10, 32)
		if err != nil {
			return nil, err
		}

		machine := &Machine{
			IP:             line[0],
			NatIP:          line[1],
			Port:           int(port),
			Username:       line[3],
			Password:       line[4],
			PrivateKeyPath: line[5],
			Device:         line[6],
		}
		machines = append(machines, machine)

		line = line[:0]
	}
	return machines, nil
}

func (m MachineList) Find(cond string) ([]*Machine, error) {
	if no, err := strconv.Atoi(cond); err == nil {
		if no <= 0 {
			goto final
		}
		if len(m) < no {
			goto final
		}
		return []*Machine{m[no-1]}, nil
	}

	if IP := net.ParseIP(cond); IP != nil {
		machines := make([]*Machine, 0, 4)
		for _, machine := range m {
			if machine.NatIP == IP.String() {
				machines = append(machines, machine)
				continue
			}
			if machine.IP == IP.String() {
				machines = append(machines, machine)
				continue
			}
		}
		if len(machines) == 0 {
			return nil, ErrNotFound
		}
		return machines, nil
	}

final:
	machines := make([]*Machine, 0, 4)
	for _, machine := range m {
		if strings.Contains(machine.IP, cond) {
			machines = append(machines, machine)
			continue
		}
		if strings.Contains(machine.NatIP, cond) {
			machines = append(machines, machine)
			continue
		}
	}
	if len(machines) == 0 {
		return nil, ErrNotFound
	}
	return machines, nil
}
