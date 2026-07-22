package runtime

import (
	"strings"
	"testing"
)

// TestDockerRunArgsWindowsPaths verifies the docker run argv carries Windows
// host paths verbatim as single argv elements (drive letter, spaces, Unicode).
// DockerRunArgs never builds a shell string, so spaces need no quoting; Docker
// Desktop parses the drive-letter colon in a bind mount specially. This test is
// portable because DockerRunArgs does no filesystem resolution.
func TestDockerRunArgsWindowsPaths(t *testing.T) {
	root := `C:\Users\Renée\My Project`
	reports := root + `\aitriage-reports`
	cache := `C:\Users\Renée\AppData\Local\aitriage\scanners`
	args := DockerRunArgs(RunSpec{
		Image:      "img",
		HostRoot:   root,
		ReportsDir: reports,
		CacheDir:   cache,
		Argv:       []string{"scan"},
	})

	wantMounts := map[string]string{
		"source (read-only)": root + ":" + containerWorkspace + ":ro",
		"reports (rw)":       reports + ":" + containerReports + ":rw",
		"cache (rw)":         cache + ":" + containerCache + ":rw",
	}
	for name, want := range wantMounts {
		if !containsExactArg(args, want) {
			t.Errorf("%s mount not present as a single argv element: %q\nargs=%v", name, want, args)
		}
	}
	// No Windows uid:gid must be forwarded, and the socket/privileged posture holds.
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "docker.sock") || strings.Contains(joined, "--privileged") {
		t.Error("must never mount the docker socket or use --privileged")
	}
}

func containsExactArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
