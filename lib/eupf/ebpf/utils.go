package ebpf

import (
	"fmt"
	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
	"unsafe"
)

// IncreaseResourceLimits https://prototype-kernel.readthedocs.io/en/latest/bpf/troubleshooting.html#memory-ulimits
func IncreaseResourceLimits() error {
	return unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	})
}

// https://man7.org/linux/man-pages/man2/bpf.2.html
// A program array map is a special kind of array map whose
// map values contain only file descriptors referring to
// other eBPF programs.  Thus, both the key_size and
// value_size must be exactly four bytes.
type BpfMapProgArrayMember struct {
	ProgramId              uint32 `json:"id"`
	ProgramRef             uint32 `json:"fd"`
	ProgramName            string `json:"name"`
	ProgramRunCount        uint32 `json:"run_count"`
	ProgramRunCountEnabled bool   `json:"run_count_enabled"`
	ProgramDuration        uint32 `json:"duration"`
	ProgramDurationEnabled bool   `json:"duration_enabled"`
}

func ListMapProgArrayContents(m *ebpf.Map) ([]BpfMapProgArrayMember, error) {
	if m.Type() != ebpf.ProgramArray {
		return nil, fmt.Errorf("map is not a program array")
	}
	var bpfMapProgArrayMember []BpfMapProgArrayMember
	var (
		key uint32
		val *ebpf.Program
	)

	iter := m.Iterate()
	for iter.Next(&key, &val) {
		programInfo, _ := val.Info()
		programID, _ := programInfo.ID()
		runCount, runCountEnabled := programInfo.RunCount()
		runDuration, runDurationEnabled := programInfo.Runtime()
		bpfMapProgArrayMember = append(bpfMapProgArrayMember,
			BpfMapProgArrayMember{
				ProgramId:              key,
				ProgramRef:             uint32(programID),
				ProgramName:            programInfo.Name,
				ProgramRunCount:        uint32(runCount),
				ProgramRunCountEnabled: runCountEnabled,
				ProgramDuration:        uint32(runDuration),
				ProgramDurationEnabled: runDurationEnabled,
			})
	}
	return bpfMapProgArrayMember, iter.Err()
}

type QerMapElement struct {
	Id           uint32 `json:"id"`
	GateStatusUL uint8  `json:"gate_status_ul"`
	GateStatusDL uint8  `json:"gate_status_dl"`
	Qfi          uint8  `json:"qfi"`
	MaxBitrateUL uint32 `json:"max_bitrate_ul"`
	MaxBitrateDL uint32 `json:"max_bitrate_dl"`
}

func ListQerMapContents(m *ebpf.Map) ([]QerMapElement, error) {
	if m.Type() != ebpf.Array {
		return nil, fmt.Errorf("map %s is not a hash", m)
	}

	contextMap := make([]QerMapElement, 0)
	mapInfo, _ := m.Info()

	var value QerInfo
	for i := uint32(0); i < mapInfo.MaxEntries; i++ {
		err := m.Lookup(i, unsafe.Pointer(&value))
		if err != nil {
			return nil, err
		}
		contextMap = append(contextMap,
			QerMapElement{
				Id:           i,
				GateStatusUL: value.GateStatusUL,
				GateStatusDL: value.GateStatusDL,
				Qfi:          value.Qfi,
				MaxBitrateUL: value.MaxBitrateUL,
				MaxBitrateDL: value.MaxBitrateDL,
			},
		)
	}

	return contextMap, nil
}
