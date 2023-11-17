package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
	"github.com/jaypipes/ghw"
	"github.com/sooslaca/smart.go"
)

func sortAttrs(attr map[uint8]smart.AtaSmartAttr) []smart.AtaSmartAttr {
	slice := make([]smart.AtaSmartAttr, 0, len(attr))
	for _, v := range attr {
		slice = append(slice, v)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Id < slice[j].Id
	})
	return slice
}

func main() {
	block, err := ghw.Block()
	if err != nil {
		panic(err)
	}
	for _, disk := range block.Disks {
		var output []string
		dev, err := smart.Open(fmt.Sprintf("/dev/%s", disk.Name))
		if err != nil { // can't open disk for SMART
			//fmt.Printf("DISK /dev/%s: %s\n", disk.Name, err)
			continue
		}
		defer dev.Close()

		switch sm := dev.(type) {
		case *smart.SataDevice:
			c, _ := sm.Identify()

			data, err := sm.ReadSMARTData()
			if err != nil {
				//fmt.Printf("ERROR: %s\n", err)
				continue
			}
			datat, err := sm.ReadSMARTThresholds()
			if err != nil {
				//fmt.Printf("ERROR: %s\n", err)
				continue
			}
			if len(data.Attrs) == 0 {
				continue
			}
			output = append(output, "┌─")
			output = append(output, fmt.Sprintf("│ Device: /dev/%s", disk.Name))
			if c != nil {
				output = append(output, fmt.Sprintf("│ Model number: %s", strings.TrimSpace(c.ModelNumber())))
				output = append(output, fmt.Sprintf("│ Serial number: %s", strings.TrimSpace(c.SerialNumber())))
				_, capacity, _, _, _ := c.Capacity()
				output = append(output, fmt.Sprintf("│ Capacity: %s", humanize.Bytes(capacity)))
			}

			headerdiv := "├─────┬──────────"
			header := "│ ID  │ Attribute"
			footerdiv := "├─────┴──────────"
			footer := "└────────────────"
			var longesattribute int
			for _, d := range sortAttrs(data.Attrs) {
				if len(d.Name) > longesattribute {
					longesattribute = len(d.Name)
				}
			}
			if longesattribute-len(" Attribute") > 0 {
				headerdiv += strings.Repeat("─", longesattribute-len("Attribute"))
				header += strings.Repeat(" ", longesattribute-len("Attribute"))
				footerdiv += strings.Repeat("─", longesattribute-len("Attribute"))
				footer += strings.Repeat("─", longesattribute-len("Attribute"))
			}
			output = append(output, fmt.Sprintf("%s─┬───────┬───────┬───────────┬──────────", headerdiv))
			output = append(output, fmt.Sprintf("%s │ Value │ Worst │ Threshold │ Raw data ", header))

			var temperatureData smart.AtaSmartAttr
			for _, d := range sortAttrs(data.Attrs) {
				output = append(output, fmt.Sprintf("│ %3d │ %-"+strconv.Itoa(longesattribute)+"s │  %03d  │  %03d  │    %03d    │ %d ", d.Id, d.Name, d.Current, d.Worst, datat.Thresholds[d.Id], d.ValueRaw))
				if d.Id == 194 {
					temperatureData = d
				}
			}
			output = append(output, fmt.Sprintf("%s─┴───────┴───────┴───────────┴──────────", footerdiv))
			generic, _ := sm.ReadGenericAttributes()
			temp, min, max, overtempCounter, err := temperatureData.ParseAsTemperature()
			if err == nil && min > 0 {
				output = append(output, fmt.Sprintf("│ Temperature: %d C (min: %d, max: %d, overtempCount: %d)", temp, min, max, overtempCounter))
			} else {
				output = append(output, fmt.Sprintf("│ Temperature: %d C", generic.Temperature))
			}
			output = append(output, fmt.Sprintf("│ Power On Hours: %d (%s)", generic.PowerOnHours, durafmt.Parse(time.Duration(generic.PowerOnHours)*time.Hour)))
			output = append(output, fmt.Sprintf("│ Power Cycles count: %d", generic.PowerCycles))
			if generic.Read > 0 {
				output = append(output, fmt.Sprintf("│ Read     block count: %d", generic.Read))
			}
			if generic.Written > 0 {
				output = append(output, fmt.Sprintf("│ Written  block count: %d", generic.Written))
			}
			output = append(output, footer)
		case *smart.ScsiDevice:
			_, _ = sm.Capacity()
		case *smart.NVMeDevice:
			_, _ = sm.ReadSMART()
		}

		// print output
		longestline := 0
		for _, x := range output {
			if utf8.RuneCountInString(x) > longestline {
				longestline = utf8.RuneCountInString(x)
			}
		}

		var closing, spacer string
		for i, x := range output {
			closing = "│"
			spacer = " "
			last, _ := utf8.DecodeLastRuneInString(x)
			if last == '─' {
				closing = "┤"
				spacer = "─"
			}
			if i == len(output)-1 {
				closing = "┘"
				spacer = "─"
			}
			if i == 0 {
				closing = "┐"
				spacer = "─"
			}
			fmt.Printf("%s%s%s\n", x, strings.Repeat(spacer, longestline-utf8.RuneCountInString(x)), closing)
		}
	}
}
