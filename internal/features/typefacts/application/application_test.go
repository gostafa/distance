// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

type errSource struct{ err error }

func (s errSource) Load(
	context.Context,
	*outbound.FactOptions,
) (string, []model.PackageExtract, error) {
	return "", nil, s.err
}

func TestCollectPropagatesLoadError(t *testing.T) {
	sentinel := errors.New("load failed")
	_, err := NewService(
		errSource{err: sentinel},
	).Collect(t.Context(), &outbound.FactOptions{})

	if !errors.Is(err, sentinel) {
		t.Fatalf("Collect error = %v, want sentinel", err)
	}
}
