package utils

import (
	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Parameter struct {
	input    string
	level    Level
	expected string
}

func TestBump(t *testing.T) {
	parameters := []Parameter{
		{"0.1.1", "patch", "0.1.2"}, {"0.1.1", "minor", "0.2.0"}, {"0.5.2", "major", "1.0.0"},
		{"0.1.1-rc.0", "patch", "0.1.2"}, {"0.1.1-rc.10", "minor", "0.2.0"}, {"0.5.2-rc.20", "major", "1.0.0"},
		{"0.1.1-rc.0", "rc", "0.1.1-rc.1"}, {"0.1.1", "rc", "0.1.1-rc.0"}, {"0.5.2-rc.20", "noop", "0.5.2-rc.20"},
	}
	assert := assert.New(t)

	for i := range parameters {
		param := parameters[i]
		v, err := semver.NewVersion(param.input)
		assert.Nil(err)
		actual, _ := Bump(*v, param.level)
		assert.Equal(param.expected, actual.String())
	}
}
