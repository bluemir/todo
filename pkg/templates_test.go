package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToArgs(t *testing.T) {
	actual := toArgs("ssh {{.addr}} -C {{command}}", map[string][]string{
		"command": []string{"cat", "abc"},
	})
	expect := []string{"ssh", "{{.addr}}", "-C", "cat", "abc"}
	assert.Equal(t, expect, actual)
}

func TestToArgs1(t *testing.T) {
	actual := toArgs("ssh {{.addr}} -C {{command}}", map[string][]string{
		"command": []string{"cat", "abc", "{{.addr}}"},
	})
	expect := []string{"ssh", "{{.addr}}", "-C", "cat", "abc", "{{.addr}}"}
	assert.Equal(t, expect, actual)
}

func TestToArgs2(t *testing.T) {
	actual := toArgs("ssh {{.addr}} -C {{command}} {{src}}", map[string][]string{
		"command": []string{"cat", "abc"},
	})
	expect := []string{"ssh", "{{.addr}}", "-C", "cat", "abc", "{{src}}"}
	assert.Equal(t, expect, actual)
}
