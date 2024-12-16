package main

import (
	"log"

	"github.com/eviltomorrow/toolbox/apps/fail2ban/cmd"
	"github.com/eviltomorrow/toolbox/lib/buildinfo"
	"github.com/eviltomorrow/toolbox/lib/system"
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
	if err := system.LoadRuntime(); err != nil {
		log.Fatalf("[F] App: load system runtime failure, nest error: %v", err)
	}

	if err := cmd.RunApp(); err != nil {
		log.Fatalf("[F] App: run app failure, nest error: %v", err)
	}
}
