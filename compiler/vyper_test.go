package compiler

import (
	"os/exec"
	"testing"
)

func TestVyper(t *testing.T) {
	if _, err := exec.LookPath("vyper"); err != nil {
		t.Skipf("Vyper compiler not installed")
	}
	v, err := NewCompiler("vyper", "vyper")
	if err != nil {
		t.Fatal(err)
	}

	output, err := v.Compile("./fixtures/auction.v.py", "./fixtures/crowdfund.v.py")
	if err != nil {
		t.Fatal(err)
	}
	if len(output) != 2 {
		t.Fatal("2 expected")
	}
}
