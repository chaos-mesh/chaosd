package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
	"github.com/chaos-mesh/chaos-mesh/pkg/ptrace"
	"github.com/chaos-mesh/chaos-mesh/pkg/time/utils"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"os"
	"runtime"
	"strings"
	"syscall"
)

// Copied from chaos-mesh/pkg/time/time_linux_amd64,
// I will move the recover part into it just future.
type ClockOption struct {
	CommonAttackConfig

	Pid           int
	SecDelta      int64
	NsecDelta     int64
	ClockIdsSlice string

	CheckPidExist bool

	ClockIdsMask uint64
}

var fakeImage = []byte{
	0xb8, 0xe4, 0x00, 0x00, 0x00, //mov    $0xe4,%eax
	0x0f, 0x05, //syscall
	0xba, 0x01, 0x00, 0x00, 0x00, //mov    $0x1,%edx
	0x89, 0xf9, //mov    %edi,%ecx
	0xd3, 0xe2, //shl    %cl,%edx
	0x48, 0x8d, 0x0d, 0x74, 0x00, 0x00, 0x00, //lea    0x74(%rip),%rcx        # <CLOCK_IDS_MASK>
	0x48, 0x63, 0xd2, //movslq %edx,%rdx
	0x48, 0x85, 0x11, //test   %rdx,(%rcx)
	0x74, 0x6b, //je     108a <clock_gettime+0x8a>
	0x48, 0x8d, 0x15, 0x6d, 0x00, 0x00, 0x00, //lea    0x6d(%rip),%rdx        # <TV_SEC_DELTA>
	0x4c, 0x8b, 0x46, 0x08, //mov    0x8(%rsi),%r8
	0x48, 0x8b, 0x0a, //mov    (%rdx),%rcx
	0x48, 0x8d, 0x15, 0x67, 0x00, 0x00, 0x00, //lea    0x67(%rip),%rdx        # <TV_NSEC_DELTA>
	0x48, 0x8b, 0x3a, //mov    (%rdx),%rdi
	0x4a, 0x8d, 0x14, 0x07, //lea    (%rdi,%r8,1),%rdx
	0x48, 0x81, 0xfa, 0x00, 0xca, 0x9a, 0x3b, //cmp    $0x3b9aca00,%rdx
	0x7e, 0x1c, //jle    <clock_gettime+0x60>
	0x0f, 0x1f, 0x40, 0x00, //nopl   0x0(%rax)
	0x48, 0x81, 0xef, 0x00, 0xca, 0x9a, 0x3b, //sub    $0x3b9aca00,%rdi
	0x48, 0x83, 0xc1, 0x01, //add    $0x1,%rcx
	0x49, 0x8d, 0x14, 0x38, //lea    (%r8,%rdi,1),%rdx
	0x48, 0x81, 0xfa, 0x00, 0xca, 0x9a, 0x3b, //cmp    $0x3b9aca00,%rdx
	0x7f, 0xe8, //jg     <clock_gettime+0x48>
	0x48, 0x85, 0xd2, //test   %rdx,%rdx
	0x79, 0x1e, //jns    <clock_gettime+0x83>
	0x4a, 0x8d, 0xbc, 0x07, 0x00, 0xca, 0x9a, //lea    0x3b9aca00(%rdi,%r8,1),%rdi
	0x3b,             //
	0x0f, 0x1f, 0x00, //nopl   (%rax)
	0x48, 0x89, 0xfa, //mov    %rdi,%rdx
	0x48, 0x83, 0xe9, 0x01, //sub    $0x1,%rcx
	0x48, 0x81, 0xc7, 0x00, 0xca, 0x9a, 0x3b, //add    $0x3b9aca00,%rdi
	0x48, 0x85, 0xd2, //test   %rdx,%rdx
	0x78, 0xed, //js     <clock_gettime+0x70>
	0x48, 0x01, 0x0e, //add    %rcx,(%rsi)
	0x48, 0x89, 0x56, 0x08, //mov    %rdx,0x8(%rsi)
	0xc3, //retq
	// constant
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //CLOCK_IDS_MASK
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TV_SEC_DELTA
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TV_NSEC_DELTA
}

func (opt *ClockOption) Attack(options AttackConfig, env chaosd.Environment) error {
	runtime.LockOSThread()
	defer func() {
		runtime.UnlockOSThread()
	}()

	program, err := ptrace.Trace(opt.Pid)
	if err != nil {
		return err
	}
	defer func() {
		err = program.Detach()
		if err != nil {
			log.Error("fail to detach program", zap.Error(err), zap.Int("pid", opt.Pid))
		}
	}()

	var vdsoEntry *mapreader.Entry
	for index := range program.Entries {
		// reverse loop is faster
		e := program.Entries[len(program.Entries)-index-1]
		if e.Path == "[vdso]" {
			vdsoEntry = &e
			break
		}
	}
	if vdsoEntry == nil {
		return fmt.Errorf("cannot find [vdso] entry")
	}

	// minus tailing variable part
	// 24 = 3 * 8 because we have three variables
	constImageLen := len(fakeImage) - 24
	var fakeEntry *mapreader.Entry

	// find injected image to avoid redundant inject (which will lead to memory leak)
	for _, e := range program.Entries {
		e := e

		image, err := program.ReadSlice(e.StartAddress, uint64(constImageLen))
		if err != nil {
			continue
		}

		if bytes.Equal(*image, fakeImage[0:constImageLen]) {
			fakeEntry = &e
			log.Error("found injected image", zap.Uint64("addr", fakeEntry.StartAddress))
			break
		}
	}

	if fakeEntry == nil {
		fakeEntry, err = program.MmapSlice(fakeImage)
		if err != nil {
			return err
		}
	}
	fakeAddr := fakeEntry.StartAddress

	// 139 is the index of CLOCK_IDS_MASK in fakeImage
	err = program.WriteUint64ToAddr(fakeAddr+139, opt.ClockIdsMask)
	if err != nil {
		return err
	}

	// 147 is the index of TV_SEC_DELTA in fakeImage
	err = program.WriteUint64ToAddr(fakeAddr+147, uint64(opt.NsecDelta))
	if err != nil {
		return err
	}

	// 155 is the index of TV_NSEC_DELTA in fakeImage
	err = program.WriteUint64ToAddr(fakeAddr+155, uint64(opt.SecDelta))
	if err != nil {
		return err
	}

	originAddr, err := program.FindSymbolInEntry("clock_gettime", vdsoEntry)
	if err != nil {
		return err
	}

	err = program.JumpToFakeFunc(originAddr, fakeAddr)
	return err
}

func (opt *ClockOption) Recover(experiment Experiment, env chaosd.Environment) error {
	panic("implement me")
}

func NewClockOption() *ClockOption {
	return &ClockOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: ClockAttack,
		},
	}
}

func (opt *ClockOption) PreProcess() error {
	clkIds := strings.Split(opt.ClockIdsSlice, ",")

	clockIdsMask, err := utils.EncodeClkIds(clkIds)
	if err != nil {
		log.Error("error while converting clock ids to mask", zap.Error(err))
		return err
	}
	opt.ClockIdsMask = clockIdsMask

	if opt.CheckPidExist {
		process, err := os.FindProcess(opt.Pid)
		if err != nil {
			fmt.Printf("Failed to find process: %s\n", err)
		} else {
			err := process.Signal(syscall.Signal(0))
			if err != nil {
				fmt.Printf("Pid may not be accessible , because : %v", err)
			}
		}
	}
	return nil
}

func (opt *ClockOption) RecoverData() string {
	data, _ := json.Marshal(opt)

	return string(data)
}
