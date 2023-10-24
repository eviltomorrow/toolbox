package system

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/eviltomorrow/toolbox/lib/netutil"
	"github.com/eviltomorrow/toolbox/lib/timeutil"
)

func init() {
	executePath, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("panic: get executable path failure, nest error: %v", err))
	}
	executePath, err = filepath.Abs(executePath)
	if err != nil {
		panic(fmt.Errorf("panic: get abs path failure, nest error: %v", err))
	}

	Runtime.ExecuteDir, Runtime.ExecuteFile = filepath.Dir(executePath), filepath.Base(executePath)
	if strings.HasSuffix(Runtime.ExecuteDir, "/bin") {
		Runtime.RootDir = filepath.Dir(Runtime.ExecuteDir)
	} else {
		Runtime.RootDir = Runtime.ExecuteDir
	}

	Runtime.HostName, _ = os.Hostname()
	Runtime.IP, _ = netutil.GetLocalIP2()
}

type runtimeHelper struct {
	ExecuteDir      string        `json:"execute-dir"`
	ExecuteFile     string        `json:"execute-file"`
	RootDir         string        `json:"root-dir"`
	Pid             int           `json:"pid"`
	LaunchTime      time.Time     `json:"launch-time"`
	HostName        string        `json:"host-name"`
	OS              string        `json:"os"`
	ARCH            string        `json:"arch"`
	RunningDuration func() string `json:"-"`
	IP              string        `json:"ip"`
}

var (
	now     = time.Now()
	Runtime = runtimeHelper{
		ARCH:       runtime.GOARCH,
		OS:         runtime.GOOS,
		Pid:        os.Getpid(),
		LaunchTime: now,
		RunningDuration: func() string {
			return timeutil.FormatDuration(time.Since(now))
		},
	}
)
