// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package profiling

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
)

func defaultProfilingRuntime() profilingRuntime {
	return profilingRuntime{
		writeHeapProfile: pprof.WriteHeapProfile,
		closeFile:        func(file *os.File) error { return file.Close() },
	}
}

// StartCPU begins CPU profiling to path and returns a stop function.
func StartCPU(path string) (stop func() error, err error) {
	stop, err = defaultProfilingRuntime().startCPU(path)
	if err != nil {
		return nil, fmt.Errorf("profiling StartCPU: %w", err)
	}

	return stop, nil
}

func (runtime profilingRuntime) startCPU(path string) (func() error, error) {
	file, err := createProfileFile(path)
	if err != nil {
		return nil, fmt.Errorf("create cpu profile: %w", err)
	}

	startErr := pprof.StartCPUProfile(file)
	if startErr != nil {
		failErr := closeStartFailed(startErr, file, runtime)

		return nil, fmt.Errorf(errWrapStartCPU, failErr)
	}

	return runtime.stopCPU(file), nil
}

func closeStartFailed(startErr error, file *os.File, runtime profilingRuntime) error {
	closeErr := wrapCloseAfter(fmt.Errorf(errWrapStartCPUProf, startErr), file, runtime)
	if closeErr != nil {
		return fmt.Errorf(errWrapStartCPU, closeErr)
	}

	return fmt.Errorf(errWrapStartCPUProf, startErr)
}

func (runtime profilingRuntime) stopCPU(file *os.File) func() error {
	return func() error {
		pprof.StopCPUProfile()

		closeErr := wrapCloseAfter(nil, file, runtime)
		if closeErr != nil {
			return fmt.Errorf("profiling stopCPU: %w", closeErr)
		}

		return nil
	}
}

// WriteHeap writes a heap profile to path.
func WriteHeap(path string) error {
	err := defaultProfilingRuntime().writeHeap(path)
	if err != nil {
		return fmt.Errorf("profiling WriteHeap: %w", err)
	}

	return nil
}

func (runtime profilingRuntime) writeHeap(path string) error {
	file, err := createProfileFile(path)
	if err != nil {
		return fmt.Errorf("create memory profile: %w", err)
	}

	closeErr := closeHeapFile(runtime, file, writeHeapTo(runtime, file))
	if closeErr != nil {
		return fmt.Errorf(errWrapWriteHeap, closeErr)
	}

	return nil
}

func closeHeapFile(runtime profilingRuntime, file *os.File, writeErr error) error {
	prior := writeErr

	if writeErr != nil {
		prior = fmt.Errorf("write memory profile: %w", writeErr)
	}

	err := finishHeap(prior, file, runtime)
	if err != nil {
		return fmt.Errorf(errWrapWriteHeap, err)
	}

	return nil
}

func finishHeap(prior error, file *os.File, closer fileCloser) error {
	closeErr := wrapCloseAfter(prior, file, closer)
	if closeErr != nil {
		return fmt.Errorf(errWrapWriteHeap, closeErr)
	}

	return nil
}

func (runtime profilingRuntime) writeTo(writer io.Writer) error {
	err := runtime.writeHeapProfile(writer)
	if err != nil {
		return fmt.Errorf("profiling writeTo: %w", err)
	}

	return nil
}

func (runtime profilingRuntime) closeNamed(file *os.File) error {
	err := runtime.closeFile(file)
	if err != nil {
		return fmt.Errorf("profiling closeNamed: %w", err)
	}

	return nil
}

func writeHeapTo(sink heapSink, writer io.Writer) error {
	err := sink.writeTo(writer)
	if err != nil {
		return fmt.Errorf("profiling writeHeapTo: %w", err)
	}

	return nil
}

func closeNamedFile(closer fileCloser, file *os.File) error {
	err := closer.closeNamed(file)
	if err != nil {
		return fmt.Errorf("profiling closeNamedFile: %w", err)
	}

	return nil
}

func wrapCloseAfter(prior error, file *os.File, closer fileCloser) error {
	err := closeAfter(prior, file, closer)
	if err != nil {
		return fmt.Errorf("profiling closeAfter: %w", err)
	}

	return nil
}

func closeAfter(prior error, file *os.File, closer fileCloser) error {
	closeErr := closeNamedFile(closer, file)

	if prior != nil {
		return prior
	}

	if closeErr != nil {
		return fmt.Errorf("close profile: %w", closeErr)
	}

	return nil
}

func createProfileFile(path string) (*os.File, error) {
	root, name, err := openProfileRoot(path)
	if err != nil {
		return nil, fmt.Errorf(errWrapOpenProfileDir, err)
	}

	file, createErr := createInRoot(root, name)
	if createErr != nil {
		return nil, fmt.Errorf(errWrapCreateProfile, createErr)
	}

	return file, nil
}

func createInRoot(root *os.Root, name string) (*os.File, error) {
	file, createErr := root.Create(name)
	closeErr := root.Close()

	if createErr != nil {
		return nil, fmt.Errorf(errWrapCreateProfile, createErr)
	}

	out, afterErr := fileAfterRootClose(file, closeErr)
	if afterErr != nil {
		return nil, fmt.Errorf(errWrapCreateProfile, afterErr)
	}

	return out, nil
}

func openProfileRoot(path string) (*os.Root, string, error) {
	dir, name := filepath.Split(path)

	if dir == emptyString {
		dir = currentDir
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, emptyString, fmt.Errorf(errWrapOpenProfileDir, err)
	}

	return root, name, nil
}

func fileAfterRootClose(file *os.File, closeErr error) (*os.File, error) {
	if closeErr == nil {
		return file, nil
	}

	discardErr := file.Close()
	if discardErr != nil {
		return nil, fmt.Errorf("close profile file: %w", discardErr)
	}

	return nil, fmt.Errorf("close profile directory: %w", closeErr)
}
