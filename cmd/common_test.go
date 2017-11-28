package cmd

import (
	"testing"
)

func TestDetectWorkEnv(t *testing.T) {
	workEnv := detectWorkEnv(".")
	println(workEnv.Branch)
}
