package runstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// atomicWrite writes data to path via a temp file + fsync + rename, so a reader
// never observes a partially written file and a crash cannot corrupt the
// bundle. Files are created with 0600 permissions.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return err
	}
	return nil
}

// readJSON reads and decodes a JSON file. The os.IsNotExist error is returned
// verbatim so callers can treat "absent" as a normal condition.
func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// dirSize returns the total size in bytes of all regular files under dir.
func dirSize(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}

// withinDir reports whether p is root itself or a descendant of root. Both must
// be cleaned, absolute, symlink-resolved paths.
func withinDir(root, p string) bool {
	if p == root {
		return true
	}
	return strings.HasPrefix(p, root+string(os.PathSeparator))
}

// skipTreeDirs are excluded from the tree fingerprint: VCS internals, the run
// bundle itself, and heavy dependency directories that are not source.
var skipTreeDirs = map[string]bool{
	".git":             true,
	".aitriage":        true, // legacy layout, still skipped if present
	"aitriage-reports": true,
	"node_modules":     true,
	"vendor":           true,
	".venv":            true,
}

// TreeFingerprint returns a stable digest of the working tree derived from
// relative paths, sizes, and modification times — never file contents. It is
// used to detect drift between approval and fix, so it does not need to be a
// content hash and stays cheap on large repositories.
func TreeFingerprint(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	var entries []string
	err = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipTreeDirs[d.Name()] && p != abs {
				return filepath.SkipDir
			}
			return nil
		}
		rel, rerr := filepath.Rel(abs, p)
		if rerr != nil {
			return rerr
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		entries = append(entries, fmt.Sprintf("%s|%d|%d", rel, info.Size(), info.ModTime().UnixNano()))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(sum[:]), nil
}
