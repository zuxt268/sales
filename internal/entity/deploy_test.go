package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTempDomains(t *testing.T) {
	assert.False(t, IsTempDomains("maruzen-ss.co.jp"))
	assert.True(t, IsTempDomains("aaa.sv533.com"))
	assert.True(t, IsTempDomains("bbb.hp-standard.com"))
}
