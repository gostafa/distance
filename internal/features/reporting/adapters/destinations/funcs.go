// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package destinations

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

// CreateFile opens a truncated output file; callers own closing it.
// Failures expose ErrOutput and diagnostic text without filesystem error chains.
func CreateFile(path string) (*os.File, error) {
	file, err := createAt(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", outbound.ErrOutput, err)
	}

	return file, nil
}

func createAt(path string) (*os.File, error) {
	dir, name := filepath.Split(path)

	if dir == "" {
		dir = "."
	}

	file, err := createNamed(dir, name)
	if err != nil {
		return nil, fmt.Errorf("create file at path: %w", err)
	}

	return file, nil
}

func createInRoot(root directory, name string) (*os.File, error) {
	file, createErr := root.Create(name)
	rootCloseErr := root.Close()

	if createErr != nil {
		return nil, errors.Join(createErr, rootCloseErr)
	}

	if rootCloseErr != nil {
		return nil, errors.Join(rootCloseErr, file.Close())
	}

	return file, nil
}

func createNamed(dir, name string) (*os.File, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open file directory: %w", err)
	}

	file, err := createInRoot(root, name)
	if err != nil {
		return nil, fmt.Errorf("create named file: %w", err)
	}

	return file, nil
}
