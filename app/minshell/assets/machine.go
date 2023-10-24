package assets

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var (
	ErrNotFound = errors.New("not found")
)

type MachineList []*Machine

type Machine struct {
	NatIP      string        `json:"nat-ip"`
	IP         string        `json:"ip"`
	Username   string        `json:"username"`
	Password   string        `json:"password"`
	Port       int           `json:"port"`
	Timeout    time.Duration `json:"timeout"`
	PrivateKey string        `json:"private-key"`
}

func LoadFile(path string) (MachineList, error) {
	if strings.HasSuffix(path, ".xlsx") {
		return loadExcelFile(path)
	}
	return nil, fmt.Errorf("not support file, path: %v", path)
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
			if colCount >= 6 {
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

		var machine = &Machine{
			IP:         line[0],
			NatIP:      line[1],
			Port:       int(port),
			Username:   line[3],
			Password:   line[4],
			PrivateKey: line[5],
		}
		machines = append(machines, machine)

		line = line[:0]
	}
	return machines, nil
}

func (m MachineList) Find(cond string) (*Machine, error) {
	if no, err := strconv.Atoi(cond); err == nil {
		if no <= 0 {
			return nil, fmt.Errorf("invalid no string")
		}
		if len(m) < no {
			return nil, ErrNotFound
		}
		return m[no-1], nil
	}

	if IP := net.ParseIP(cond); IP != nil {
		for _, machine := range m {
			if machine.NatIP == IP.String() {
				return machine, nil
			}
			if machine.IP == IP.String() {
				return machine, nil
			}
		}
		return nil, ErrNotFound
	}
	return nil, fmt.Errorf("invalid input string")
}
