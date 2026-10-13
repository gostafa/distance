// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"errors"
	"io"
	"testing"

	"github.com/gostafa/distance/internal/features/reporting/domain"
)

type fakeOutput struct {
	err          error
	writes, docs int
}

func (f *fakeOutput) Write(
	report *domain.Report,
	w io.WriteCloser,
	opts *domain.WriteOptions,
) error {
	f.writes++
	return f.err
}

func (f *fakeOutput) WriteDocs(w io.WriteCloser, version string) error { f.docs++; return f.err }

func TestOutputPortSuccessAndFailure(t *testing.T) {
	fake := &fakeOutput{}
	service := New(fake)
	if err := service.Write(
		&domain.Report{},
		nil,
		&domain.WriteOptions{Format: "json"},
	); err != nil {
		t.Fatal(err)
	}
	if err := service.WriteDocs(nil, "dev"); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("output failed")
	fake.err = failure
	if err := service.Write(
		&domain.Report{},
		nil,
		&domain.WriteOptions{},
	); !errors.Is(
		err,
		failure,
	) {
		t.Fatalf("write failure lost: %v", err)
	}
	if err := service.WriteDocs(nil, "dev"); !errors.Is(err, failure) {
		t.Fatalf("guide failure lost: %v", err)
	}
	if fake.writes != 2 || fake.docs != 2 {
		t.Fatalf("calls: %+v", fake)
	}
}
