package application_test

import (
	"context"
	"errors"
	"testing"

	projectanalysis "github.com/gostafa/distance/internal/features/projectanalysis/application"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

type errSource struct{ err error }

func (e errSource) Load(
	context.Context,
	tfoutbound.FactOptions,
) (string, []tfdomain.PackageExtract, error) {
	return "", nil, e.err
}

// Black-box: a fact-source failure propagates out of Analyze.
func TestPipelineLoadError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("load failed")
	pipeline := projectanalysis.NewPipeline(typefacts.NewService(errSource{err: sentinel}))

	_, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns: []string{"./..."},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel", err)
	}
}
