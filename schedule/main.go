package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ====== CPU Affinity (Linux sched_setaffinity) ======

const (
	_CPU_SETSIZE  = 1024
	_CPU_SETBYTES = _CPU_SETSIZE / 8
)

// cpuSet is a bitset compatible with Linux cpu_set_t (up to 1024 CPUs)
type cpuSet [_CPU_SETBYTES]byte

func (c *cpuSet) set(cpu int) {
	c[cpu/8] |= 1 << (uint(cpu) % 8)
}

func schedSetAffinitySelf(cpus []int) error {
	var set cpuSet
	for _, c := range cpus {
		if c < 0 || c >= _CPU_SETSIZE {
			return fmt.Errorf("cpu id out of range: %d", c)
		}
		set.set(c)
	}
	_, _, errno := syscall.RawSyscall(
		syscall.SYS_SCHED_SETAFFINITY,
		0, // pid 0 = current thread
		uintptr(len(set)),
		uintptr(unsafe.Pointer(&set)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// ====== cpuset parsing (cgroup v1) ======

func readCpusetCpus() ([]int, error) {
	// 常见路径：/sys/fs/cgroup/cpuset/cpuset.cpus 或 /sys/fs/cgroup/cpuset/cpuset.cpus.effective
	paths := []string{
		"/sys/fs/cgroup/cpuset/cpuset.cpus.effective",
		"/sys/fs/cgroup/cpuset/cpuset.cpus",
	}
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil && len(strings.TrimSpace(string(data))) > 0 {
			return parseCPUList(string(data))
		}
	}
	return nil, errors.New("cpuset.cpus not found or empty")
}

func parseCPUList(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty cpuset")
	}
	var cpus []int
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			ab := strings.SplitN(part, "-", 2)
			a, _ := strconv.Atoi(ab[0])
			b, _ := strconv.Atoi(ab[1])
			for i := a; i <= b; i++ {
				cpus = append(cpus, i)
			}
		} else {
			n, _ := strconv.Atoi(part)
			cpus = append(cpus, n)
		}
	}
	return cpus, nil
}

// ====== Example workload ======

func matcherLoop(id int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := schedSetAffinitySelf([]int{id}); err != nil {
		fmt.Println("matcher affinity error:", err)
		return
	}
	// 撮合主循环（示例）
	for {
		// do matching
		time.Sleep(1 * time.Millisecond)
	}
}

func walLoop(cpu int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := schedSetAffinitySelf([]int{cpu}); err != nil {
		fmt.Println("wal affinity error:", err)
		return
	}
	// WAL 刷盘线程（示例）
	t := time.NewTicker(5 * time.Millisecond)
	defer t.Stop()
	for range t.C {
		// flush WAL
	}
}

func main() {
	allowed, err := readCpusetCpus()
	if err != nil {
		fmt.Println("cannot read cpuset, fallback to all CPUs:", err)
		allowed = make([]int, runtime.NumCPU())
		for i := range allowed {
			allowed[i] = i
		}
	}

	if len(allowed) < 2 {
		fmt.Println("need at least 2 CPUs for pinning strategy")
		return
	}

	// 预留一个 CPU 给 runtime/IO
	runtimeCPU := allowed[0]
	matcherCPUs := allowed[1:]

	// 建议 GOMAXPROCS = len(allowed)
	runtime.GOMAXPROCS(len(allowed))

	// 绑定 WAL 到 runtimeCPU（或者最后一个核，看你喜好）
	go walLoop(runtimeCPU)

	// 启动 matcher shard（一个 goroutine 绑一个核）
	for _, c := range matcherCPUs {
		go matcherLoop(c)
	}

	// 主线程也可绑定到 runtimeCPU
	runtime.LockOSThread()
	_ = schedSetAffinitySelf([]int{runtimeCPU})
	runtime.UnlockOSThread()

	// 阻塞
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}
