//go:build linux

package fs

import (
	"os"
	"syscall"
)

var Ctime = func(fi os.FileInfo) int64 {
	stat := fi.Sys().(*syscall.Stat_t)

	return (stat.Ctim.Sec*1_000_000_000 + stat.Ctim.Nsec) / 1000 // Look for CONFIG_HZ in README.md
}
