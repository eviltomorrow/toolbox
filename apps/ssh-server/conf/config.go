package conf

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/eviltomorrow/toolbox/lib/flagsutil"
	"github.com/eviltomorrow/toolbox/lib/system"
	jsoniter "github.com/json-iterator/go"
)

type Config struct {
	Server Server          `json:"server" toml:"server" mapstructure:"server"`
	Users  map[string]User `json:"users" toml:"users" mapstructure:"users"`
	Log    Log             `json:"log" toml:"log" mapstructure:"log"`
}

type Server struct {
	BlackList         []string `json:"black-list" toml:"black-list" mapstructure:"black-list"`
	Port              int      `json:"port" toml:"port" mapstructure:"port"`
	PrivateKey        string   `json:"private-key" toml:"private-key" mapstructure:"private-key"`
	MaximumLoginLimit int      `json:"maximum-login-limit" toml:"maximum-login-limit" mapstructure:"maximum-login-limit"`
}

type User struct {
	Username string `json:"username" toml:"username" mapstructure:"username"`
	Password string `json:"password" toml:"password" mapstructure:"password"`
}

type Log struct {
	Level         string `json:"level" toml:"level" mapstructure:"level"`
	DisableStdlog bool   `json:"disable-stdlog" toml:"-" mapstructure:"-"`
}

func (c *Config) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(c)
	return string(buf)
}

func ReadConfig(opts *flagsutil.Flags) (*Config, error) {
	path := func() string {
		if opts.ConfigFile != "" {
			return opts.ConfigFile
		}
		return filepath.Join(system.Directory.EtcDir, "config.toml")
	}()

	if _, err := toml.DecodeFile(path, &DefaultConfig); err != nil {
		return nil, err
	}

	if err := DefaultConfig.isConfigValid(); err != nil {
		return nil, fmt.Errorf("invalid config, nest error: %v", err)
	}

	DefaultConfig.Log.DisableStdlog = opts.DisableStdlog

	return DefaultConfig, nil
}

func (c *Config) isConfigValid() error {
	if c.Server.Port <= 0 || c.Server.Port >= 65535 {
		return fmt.Errorf("invalid server.port[%d]", c.Server.Port)
	}
	if c.Server.PrivateKey == "" {
		return fmt.Errorf("server.private-key is nil")
	}
	if len(c.Users) == 0 {
		return fmt.Errorf("users is nil")
	}

	for key, user := range c.Users {
		if user.Username == "" {
			return fmt.Errorf("users.%s username is nil", key)
		}
		if user.Password == "" {
			return fmt.Errorf("users.%s password is nil", key)
		}
	}
	return nil
}

var DefaultConfig = &Config{
	Server: Server{
		BlackList:         []string{},
		Port:              18080,
		PrivateKey:        "./etc/id_ed25519",
		MaximumLoginLimit: 10,
	},
	Users: map[string]User{},
	Log: Log{
		Level:         "info",
		DisableStdlog: true,
	},
}
