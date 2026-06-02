package gotestrunner

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CleanStaleOverlays removes temporary overlay directories whose owning
// process is no longer alive, and evicts aged-out cache entries.
func CleanStaleOverlays() {
	cleanStaleTmpOverlays()
	cleanOldCacheEntries()
}

func cleanStaleTmpOverlays() {
	pattern := filepath.Join(os.TempDir(), "gotest-overlay-*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		if isOverlayAlive(dir) {
			continue
		}
		os.RemoveAll(dir)
	}
}

const cacheMaxAge = 7 * 24 * time.Hour

func cleanOldCacheEntries() {
	root, err := cacheRoot()
	if err != nil {
		return
	}
	overlaysDir := filepath.Join(root, "overlays")
	entries, err := os.ReadDir(overlaysDir)
	if err != nil {
		return
	}
	now := time.Now()
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > cacheMaxAge {
			os.RemoveAll(filepath.Join(overlaysDir, e.Name()))
		}
	}
}

func isOverlayAlive(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, ".pid"))
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	return processAlive(pid)
}
