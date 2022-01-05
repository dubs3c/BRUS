//go:build linux

package main

import (
	"io/fs"
	"syscall"
	"time"
)

func GetCreationDate(file fs.FileInfo) time.Time {
	ts := file.Sys().(*syscall.Stat_t).Mtim
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}
