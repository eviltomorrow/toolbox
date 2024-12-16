package conf

import (
	"testing"

	"github.com/eviltomorrow/toolbox/lib/flagsutil"
	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	assert := assert.New(t)
	config, err := ReadConfig(&flagsutil.Flags{ConfigFile: "etc/fail2ban.yml"})
	assert.Nil(err)
	assert.Equal(1, len(config.Rules))
	t.Logf("%v", config)
}
