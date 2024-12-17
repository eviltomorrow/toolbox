package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/eviltomorrow/toolbox/apps/ssh-server/conf"
	"github.com/eviltomorrow/toolbox/lib/buildinfo"
	"github.com/eviltomorrow/toolbox/lib/finalizer"
	"github.com/eviltomorrow/toolbox/lib/flagsutil"
	"github.com/eviltomorrow/toolbox/lib/fs"
	"github.com/eviltomorrow/toolbox/lib/pprofutil"
	"github.com/eviltomorrow/toolbox/lib/procutil"
	"github.com/eviltomorrow/toolbox/lib/system"
	"github.com/eviltomorrow/toolbox/lib/zlog"
	flags "github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

func RunApp() error {
	_, err := flags.Parse(flagsutil.Opts)
	if err != nil {
		return err
	}

	if flagsutil.Opts.Version {
		fmt.Println(buildinfo.Version())
		os.Exit(0)
	}

	if flagsutil.Opts.Daemon {
		if err := procutil.RunAppInBackground(os.Args); err != nil {
			log.Fatalf("[F] Daemon: run app in background failure, nest error: %v", err)
		}
		return nil
	}

	if flagsutil.Opts.EnablePprof {
		go func() {
			if err := pprofutil.Run(flagsutil.Opts.PprofAddr); err != nil {
				log.Fatalf("[F] Run pprof failure, nest error: %v", err)
			}
		}()
	}
	defer func() {
		finalizer.RunCleanupFuncs()
	}()

	c, err := conf.ReadConfig(flagsutil.Opts)
	if err != nil {
		return fmt.Errorf("read config failure, nest error: %v", err)
	}

	releaseFile, err := procutil.CreatePidFile()
	if err != nil {
		return fmt.Errorf("create pid file failure, nest error ;%v", err)
	}
	finalizer.RegisterCleanupFuncs(releaseFile)

	// 必须在最后
	if err := fs.RewriteStderrToFile(); err != nil {
		return fmt.Errorf("rewrite stderr to filre failure, nest error: %v", err)
	}

	zlog.Info("System info", zap.String("system", system.String()))
	zlog.Info("Config info", zap.String("config", c.String()))
	zlog.Info("App start success", zap.String("version", buildinfo.MainVersion), zap.String("commited-id", buildinfo.GitSha))

	procutil.StopDaemon()
	procutil.WaitForSigterm()

	zlog.Info("App stop complete", zap.String("launched-time", system.LaunchTime()))
	return nil
}

func InitLog() error {
	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:  "info",
		Format: "json",
		File: zlog.FileLogConfig{
			Filename:    filepath.Join(system.Directory.LogDir, "data.log"),
			MaxSize:     100,
			MaxDays:     30,
			MaxBackups:  90,
			Compression: "gzip",
		},
		DisableStacktrace: true,
		DisableStdlog:     true,
	})
	if err != nil {
		return fmt.Errorf("init global log failure, nest error: %v", err)
	}
	zlog.ReplaceGlobals(global, prop)
	finalizer.RegisterCleanupFuncs(global.Sync)

	return nil
}
