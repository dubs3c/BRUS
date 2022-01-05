//go:build windows

package main

func GetCreationDate(file fs.FileInfo) time.Time {
	ts := file.Sys().(*syscall.Win32FileAttributeData).CreationTime
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}
