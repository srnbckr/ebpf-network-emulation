// Code generated by bpf2go; DO NOT EDIT.
//go:build arm64be || armbe || mips || mips64 || mips64p32 || ppc64 || s390 || s390x || sparc || sparc64
// +build arm64be armbe mips mips64 mips64p32 ppc64 s390 s390x sparc sparc64

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

type tcHandleBpsDelay struct {
	TcHandle        uint32
	ThrottleRateBps uint32
	DelayMs         uint32
}

// loadTc returns the embedded CollectionSpec for tc.
func loadTc() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_TcBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load tc: %w", err)
	}

	return spec, err
}

// loadTcObjects loads tc and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//     *tcObjects
//     *tcPrograms
//     *tcMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadTcObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadTc()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// tcSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type tcSpecs struct {
	tcProgramSpecs
	tcMapSpecs
}

// tcSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type tcProgramSpecs struct {
	TcMain *ebpf.ProgramSpec `ebpf:"tc_main"`
}

// tcMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type tcMapSpecs struct {
	IP_HANDLE_BPS_DELAY *ebpf.MapSpec `ebpf:"IP_HANDLE_BPS_DELAY"`
}

// tcObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadTcObjects or ebpf.CollectionSpec.LoadAndAssign.
type tcObjects struct {
	tcPrograms
	tcMaps
}

func (o *tcObjects) Close() error {
	return _TcClose(
		&o.tcPrograms,
		&o.tcMaps,
	)
}

// tcMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadTcObjects or ebpf.CollectionSpec.LoadAndAssign.
type tcMaps struct {
	IP_HANDLE_BPS_DELAY *ebpf.Map `ebpf:"IP_HANDLE_BPS_DELAY"`
}

func (m *tcMaps) Close() error {
	return _TcClose(
		m.IP_HANDLE_BPS_DELAY,
	)
}

// tcPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadTcObjects or ebpf.CollectionSpec.LoadAndAssign.
type tcPrograms struct {
	TcMain *ebpf.Program `ebpf:"tc_main"`
}

func (p *tcPrograms) Close() error {
	return _TcClose(
		p.TcMain,
	)
}

func _TcClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//go:embed tc_bpfeb.o
var _TcBytes []byte