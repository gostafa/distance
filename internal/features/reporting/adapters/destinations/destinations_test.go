// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package destinations

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

type testDirectory struct {
	file                *os.File
	createErr, closeErr error
}

func (d testDirectory) Create(string) (*os.File, error) { return d.file, d.createErr }

func (d testDirectory) Close() error { return d.closeErr }

func TestFilesystemFailuresAreStable(t *testing.T) {
	for _, path := range []string{filepath.Join(t.TempDir(), "missing", "report"), t.TempDir()} {
		file, err := CreateFile(path)
		if file != nil || !errors.Is(err, outbound.ErrOutput) || errors.Is(err, os.ErrNotExist) {
			t.Fatalf("filesystem error leaked: file=%v err=%v", file, err)
		}
	}
	path := filepath.Join(t.TempDir(), "report")
	file, err := CreateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("report"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "report" {
		t.Fatalf("output contents: %q err=%v", data, err)
	}
	t.Chdir(t.TempDir())
	relative, err := CreateFile("report")
	if err != nil {
		t.Fatal(err)
	}
	if err := relative.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDirectoryLifecycleFailures(t *testing.T) {
	failure := errors.New("directory close failed")
	if _, err := createInRoot(
		testDirectory{createErr: failure},
		"report",
	); !errors.Is(
		err,
		failure,
	) {
		t.Fatalf("creation failure = %v", err)
	}
	file, err := os.CreateTemp(t.TempDir(), "report")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := createInRoot(
		testDirectory{file: file, closeErr: failure},
		"report",
	); !errors.Is(
		err,
		failure,
	) {
		t.Fatalf("close failure = %v", err)
	}
	if _, err := file.WriteString("x"); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("file was not discarded: %v", err)
	}
}
