package system

import (
	"bytes"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	startup = time.Now()
)

var (
	LaunchTime = func() string {
		return FormatDuration(time.Since(startup))
	}
	Machine   machine
	Network   network
	Process   process
	Directory directory
)

type machine struct {
	Hostname string
}

type network struct {
	AccessIP string
	BindIP   string
}

type process struct {
	Name string
	Args []string

	Pid  int
	PPid int
}

type directory struct {
	RootDir string

	BinDir string
	EtcDir string
	UsrDir string
	VarDir string
	LogDir string
}

func String() string {
	var data = map[string]interface{}{
		"machine": Machine,
		"network": Network,
		"process": Process,
	}

	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(data)
	return string(buf)
}

func FormatDuration(d time.Duration) string {
	var (
		day, hour, minute, second int
	)

	switch {
	case d.Hours() > 23.0:
		var h = int(d.Hours())
		day = h / 24
		hour = h % 24
		minute = int(d.Minutes()) - (day*24+hour)*60
		second = int(d.Seconds()) - ((day*24+hour)*60+minute)*60
	case d.Minutes() > 59.0:
		var m = int(d.Minutes())
		hour = m / 60
		minute = m % 60
		second = int(d.Seconds()) - (hour*60+minute)*60
	case d.Seconds() > 59:
		var s = int(d.Seconds())
		minute = s / 60
		second = s % 60
	default:
		second = int(d.Seconds())
	}

	var buf bytes.Buffer
	if day > 0 {
		buf.WriteString(fmt.Sprintf("%d天", day))
	}
	if hour > 0 {
		buf.WriteString(fmt.Sprintf("%d小时", hour))
	}
	if minute > 0 {
		buf.WriteString(fmt.Sprintf("%d分钟", minute))
	}
	if second > 0 {
		buf.WriteString(fmt.Sprintf("%d秒", second))
	}

	if buf.Len() == 0 {
		return "0秒"
	}
	return buf.String()
}
