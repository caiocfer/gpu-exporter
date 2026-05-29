package proc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var Root = "/proc"

func ReadComm(pid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("%s/%d/comm", Root, pid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func ReadCmdline(pid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("%s/%d/cmdline", Root, pid))
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(strings.TrimRight(string(data), "\x00"), "\x00", " ")
}

func ReadStat(pid uint32) (cpuTicks uint64, err error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/%d/stat", Root, pid))
	if err != nil {
		return 0, err
	}
	s := string(data)
	ri := strings.LastIndex(s, ")")
	if ri == -1 {
		return 0, fmt.Errorf("no closing paren in stat for pid %d", pid)
	}
	rest := strings.Fields(s[ri+2:])
	// utime = field 11, stime = field 12 after ") "
	if len(rest) < 13 {
		return 0, fmt.Errorf("unexpected stat format for pid %d", pid)
	}
	utime, err := strconv.ParseUint(rest[11], 10, 64)
	if err != nil {
		return 0, err
	}
	stime, err := strconv.ParseUint(rest[12], 10, 64)
	if err != nil {
		return 0, err
	}
	return utime + stime, nil
}

func ReadRSS(pid uint32) uint64 {
	data, err := os.ReadFile(fmt.Sprintf("%s/%d/status", Root, pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0
				}
				return kb * 1024
			}
		}
	}
	return 0
}
