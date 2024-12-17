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
	Auth   map[string]User `json:"auth" toml:"auth" mapstructure:"auth"`
}

type Server struct {
	BlackList  []string `json:"black-list" toml:"black-list" mapstructure:"black-list"`
	Port       int      `json:"port" toml:"port" mapstructure:"port"`
	PrivateKey string   `json:"private-key" toml:"private-key" mapstructure:"private-key"`
}

type User struct {
	Username string `json:"username" toml:"username" mapstructure:"username"`
	Password string `json:"password" toml:"password" mapstructure:"password"`
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

	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, err
	}

	if err := c.isConfigValid(); err != nil {
		return nil, fmt.Errorf("invalid config, nest error: %v", err)
	}
	return &c, nil
}

func (c *Config) isConfigValid() error {
	if c.Server.Port <= 0 || c.Server.Port >= 65535 {
		return fmt.Errorf("invalid server.port[%d]", c.Server.Port)
	}
	if c.Server.PrivateKey == "" {
		return fmt.Errorf("server.private-key is nil")
	}
	if len(c.Auth) == 0 {
		return fmt.Errorf("auth'user is nil")
	}
	return nil
}
