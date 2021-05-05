/*
history:
20/1030 v1
20/1106 suffif every line with shortened hostname

GoFmt GoBuildNull GoBuild GoRelease GoRun
*/

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	pscpu "github.com/shirou/gopsutil/cpu"
	psdisk "github.com/shirou/gopsutil/disk"
	pshost "github.com/shirou/gopsutil/host"
	psmem "github.com/shirou/gopsutil/mem"
)

const VisualRatio = 5

var (
	Hostname     string
	PollInterval time.Duration
	TimeLimit    time.Duration
)

func log(msg string, args ...interface{}) {
	const Beat = time.Duration(24) * time.Hour / 1000
	tzBiel := time.FixedZone("Biel", 60*60)
	t := time.Now().In(tzBiel)
	ty := t.Sub(time.Date(t.Year(), 1, 1, 0, 0, 0, 0, tzBiel))
	td := t.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tzBiel))
	ts := fmt.Sprintf(
		"%d/%d@%d",
		t.Year()%1000,
		int(ty/(time.Duration(24)*time.Hour))+1,
		int(td/Beat),
	)
	fmt.Fprintf(os.Stderr, ts+" "+msg+"\n", args...)
}

func print() {
	const Beat = time.Duration(24) * time.Hour / 1000
	tzBiel := time.FixedZone("Biel", 60*60)
	tn := time.Now().In(tzBiel)
	td := tn.Sub(time.Date(tn.Year(), tn.Month(), tn.Day(), 0, 0, 0, 0, tzBiel))
	tb := int(td / Beat)

	cpuInterval := PollInterval
	if cpuInterval == 0 {
		cpuInterval = time.Second / 10
	}
	cpupercents, err := pscpu.Percent(cpuInterval, false)
	if err != nil {
		log("pscpu.Percent: %v", err)
		os.Exit(1)
	}
	cpupercent := int(cpupercents[0])
	cpugauge := (strings.Repeat("=", cpupercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-cpupercent/VisualRatio))
	cpunumber, err := pscpu.Counts(false)
	if err != nil {
		log("pscpu.Counts: %v", err)
		os.Exit(1)
	}

	mem, err := psmem.VirtualMemory()
	if err != nil {
		log("psmem.VirtualMemory: %v", err)
		os.Exit(1)
	}
	memsizegb := int(mem.Total / (1 << 30))
	mempercent := int(mem.UsedPercent)
	memgauge := (strings.Repeat("=", mempercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-mempercent/VisualRatio))

	swap, err := psmem.SwapMemory()
	if err != nil {
		log("psmem.SwapMemory: %v", err)
		os.Exit(1)
	}
	swapsizegb := int(swap.Total / (1 << 30))
	swappercent := int(swap.UsedPercent)
	var swapgauge string
	if swap.Total > 0 {
		swapgauge = (strings.Repeat("=", swappercent/VisualRatio) +
			strings.Repeat("-", 100/VisualRatio-swappercent/VisualRatio))
	} else {
		swapgauge = strings.Repeat(" ", 100/VisualRatio)
	}

	disk, err := psdisk.Usage("/")
	if err != nil {
		log("psdisk.Usage: %v", err)
		os.Exit(1)
	}
	disksizegb := int(disk.Total / (1 << 30))
	diskpercent := int(disk.UsedPercent)
	diskgauge := (strings.Repeat("=", diskpercent/VisualRatio) +
		strings.Repeat("-", 100/VisualRatio-diskpercent/VisualRatio))

	uptimes, err := pshost.Uptime()
	if err != nil {
		log("pshost.Uptime: %v", err)
		os.Exit(1)
	}
	uptime := time.Duration(uptimes) * time.Second
	uptimedays, uptime := uptime/(24*time.Hour), uptime%(24*time.Hour)
	uptimeweeks, uptimedays := uptimedays/7, uptimedays%7
	uptimeweeksdays := ""
	if uptimeweeks > 0 {
		uptimeweeksdays += fmt.Sprintf("%dw", uptimeweeks)
	}
	if uptimedays > 0 {
		uptimeweeksdays += fmt.Sprintf("%dd", uptimedays)
	}
	if uptimeweeksdays != "" {
		uptimeweeksdays += "."
	}

	fmt.Printf(
		"@%3d\t%s\tcpu%s%d mem%s%dgb swap%s%dgb disk%s%dgb uptime:%s%s\n",
		tb, Hostname,
		cpugauge, cpunumber,
		memgauge, memsizegb,
		swapgauge, swapsizegb,
		diskgauge, disksizegb,
		uptimeweeksdays, uptime,
	)
}

func main() {
	var err error

	Hostname, err = os.Hostname()
	if err != nil {
		log("Hostname: %v", err)
		os.Exit(1)
	}
	Hostname = strings.TrimSuffix(Hostname, ".local")
	const hostnamemaxlen = 14
	if len(Hostname) > hostnamemaxlen {
		Hostname = Hostname[:hostnamemaxlen-5] + "~" + Hostname[len(Hostname)-4:]
	}

	if len(os.Args) > 1 {
		ri, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log("invalid integer `%s` for repeat interval in seconds", os.Args[1])
			os.Exit(1)
		}
		PollInterval = time.Duration(ri) * time.Second

		if len(os.Args) > 2 {
			tl, err := strconv.Atoi(os.Args[2])
			if err != nil {
				log("invalid integer `%s` for time limit in seconds", os.Args[2])
				os.Exit(1)
			}
			TimeLimit = time.Duration(tl) * time.Second
		}
	}

	if PollInterval > 0 {
		st := time.Now()
		for {
			print()
			time.Sleep(PollInterval)
			if TimeLimit > 0 && time.Since(st) > TimeLimit {
				break
			}
		}
	} else {
		print()
	}
}
