package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eviltomorrow/toolbox/lib/flagsutil"
	"github.com/eviltomorrow/toolbox/lib/system"
	"github.com/eviltomorrow/toolbox/lib/timeutil"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Interface string   `json:"interface"`
	Rules     []Rule   `json:"rules"`
	WhiteList []string `json:"whitelist"`
}

type Rule struct {
	Name     string            `json:"name"`
	Port     int               `json:"port"`
	Duration timeutil.Duration `json:"duration"`
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
		return filepath.Join(system.Directory.EtcDir, "fail2ban.yml")
	}()

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(buf, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) IsValid() error {
	if len(c.Rules) == 0 {
		return fmt.Errorf("no valid rules")
	}

	for i, rule := range c.Rules {
		if rule.Name == "" {
			return fmt.Errorf("invalid value, rules[%d]'s name is nil", i+1)
		}
		if rule.Port <= 0 || rule.Port >= 65535 {
			return fmt.Errorf("invalid value, rules[%d]'s port is invalid", i+1)
		}
		if rule.Duration.String() == "" {
			return fmt.Errorf("invalid value, rules[%d]'s duration is nil", i+1)
		}
	}
	return nil
}
