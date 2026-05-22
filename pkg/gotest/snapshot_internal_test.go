package gotest //nolint:stdlib-test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func thisDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func TestIsExternalPackage(t *testing.T) {
	dir := thisDir()

	t.Run("ptest file returns false", func(t *testing.T) {
		pkgCache.Delete(filepath.Join(dir, "collecting_test.go"))
		got := isExternalPackage(filepath.Join(dir, "collecting_test.go"))
		if got {
			t.Fatal("expected false for ptest file")
		}
	})

	t.Run("pxtest file returns true", func(t *testing.T) {
		pkgCache.Delete(filepath.Join(dir, "snapshot_test.go"))
		got := isExternalPackage(filepath.Join(dir, "snapshot_test.go"))
		if !got {
			t.Fatal("expected true for pxtest file")
		}
	})

	t.Run("nonexistent file returns false", func(t *testing.T) {
		got := isExternalPackage(filepath.Join(dir, "nonexistent.go"))
		if got {
			t.Fatal("expected false for nonexistent file")
		}
	})

	t.Run("result is cached", func(t *testing.T) {
		path := filepath.Join(dir, "collecting_test.go")
		pkgCache.Delete(path)
		isExternalPackage(path)
		_, ok := pkgCache.Load(path)
		if !ok {
			t.Fatal("expected result to be cached")
		}
	})
}

func TestMatchSnapshot_PtestUsesNoSuffix(t *testing.T) {
	snapDir := filepath.Join(thisDir(), "testdata", "__snapshots__")
	t.Cleanup(func() { os.RemoveAll(snapDir) })

	MatchSnapshot(t, "ptest-value")

	snapPath := filepath.Join(snapDir, "TestMatchSnapshot_PtestUsesNoSuffix.snap")
	data, err := os.ReadFile(snapPath)
	if err != nil {
		t.Fatalf("expected .snap (no _ext suffix): %v", err)
	}
	if !strings.Contains(string(data), "ptest-value") {
		t.Fatal("expected snapshot content")
	}

	extPath := filepath.Join(snapDir, "TestMatchSnapshot_PtestUsesNoSuffix_ext.snap")
	if _, err := os.Stat(extPath); err == nil {
		t.Fatal("_ext.snap should not exist for ptest caller")
	}
}
