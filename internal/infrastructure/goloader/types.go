// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"context"
	"go/token"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"golang.org/x/tools/go/packages"
)

type (
	// Loaded is the package-extract list returned by Loader.Load.
	Loaded = []model.PackageExtract

	pkgExtracts = []model.PackageExtract
	loadedPkgs  = []*packages.Package
	pkgByPath   = map[string]*packages.Package
	factOpts    = outbound.FactOptions

	extractorOptions struct {
		modulePath       string
		includeGenerated bool
	}

	// Loader is an outbound.FactSource backed by golang.org/x/tools/go/packages.
	Loader struct{}

	skipFilter struct {
		fset      *token.FileSet
		generated map[string]bool
		opts      *extractorOptions
	}

	extractJob struct {
		opts       *outbound.FactOptions
		modulePath string
		pkgs       []*packages.Package
	}

	filterOut struct {
		pkgs []*packages.Package
		errs []string
	}

	loadOut struct {
		err error
		mod string
		ext Loaded
	}

	loaderRuntime struct {
		packagesLoad      func(*packages.Config, ...string) ([]*packages.Package, error)
		runExtractWorkers func(context.Context, int, int, func(int) error) error
	}

	workerConfig struct {
		workers int
		tasks   int
	}
)
