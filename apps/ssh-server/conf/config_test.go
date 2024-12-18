package conf

import (
	"testing"

	"github.com/eviltomorrow/toolbox/lib/flagsutil"
	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	assert := assert.New(t)
	c, err := ReadConfig(&flagsutil.Flags{
		ConfigFile: "etc/config.toml",
	})

	assert.Nil(err)
	t.Logf("%v", c.String())
}
