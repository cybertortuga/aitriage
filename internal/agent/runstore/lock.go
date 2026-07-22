package runstore

import (
	"os"
	"path/filepath"
)

// lockRun acquires an exclusive, advisory, cross-process lock for a run bundle.
// Every mutating operation holds it while it re-reads the manifest from disk,
// applies its change, and writes back, so two Store.Open handles — in the same
// process or across processes — cannot interleave a read-modify-write on the
// same run. The returned function releases the lock.
//
// The lock file lives inside the run directory (0600) and is never part of the
// portable bundle contents (callers do not serialize or ship it).
func lockRun(dir string) (func(), error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, ".lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() {
		_ = unlockFile(f)
		_ = f.Close()
	}, nil
}
