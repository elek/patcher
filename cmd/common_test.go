package cmd

import (
	"testing"
)

func TestDetectWorkEnv(t *testing.T) {
	branch, _ := detectWorkEnv(".")
	println(branch)
}
