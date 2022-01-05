package main

import (
	"testing"
)

func TestParseLogFiles(t *testing.T) {
	// need to update creation date of test file everytime this code runs :(
	ips, err := ParseLogFiles("./test", 30)
	if err != nil {
		t.Fatal(err)
	}

	l := len(ips)
	if l != 18 {
		t.Errorf("Amount of IPs is incorrect, expected %d, got %d", 18, l)
	}
}
