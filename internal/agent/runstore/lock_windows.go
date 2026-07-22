//go:build windows

package runstore

import (
	"os"

	"golang.org/x/sys/windows"
)

// Lock the first byte of the per-run lock file. The file is opened without
// asynchronous I/O, so LockFileEx blocks until the exclusive lock is acquired,
// matching flock(LOCK_EX) on Unix.
func lockFile(f *os.File) error {
	return windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		1,
		0,
		&windows.Overlapped{},
	)
}

func unlockFile(f *os.File) error {
	return windows.UnlockFileEx(
		windows.Handle(f.Fd()),
		0,
		1,
		0,
		&windows.Overlapped{},
	)
}
