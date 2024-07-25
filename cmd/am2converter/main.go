package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/walterwanderley/am2manager"
)

var (
	level   uint
	mix     uint
	gainMin uint
	gainMax uint
)

func main() {
	flag.UintVar(&level, "level", 100, "Set the Level (0-255)")
	flag.UintVar(&mix, "mix", 100, "Set the Mix (0-100)")
	flag.UintVar(&gainMin, "gain-min", 30, "Set the minimum gain (0-100)")
	flag.UintVar(&gainMax, "gain-max", 60, "Set the maximum gain (0-100)")
	flag.Parse()

	if err := validate(); err != nil {
		slog.Error("input validation", "error", err.Error())
		os.Exit(-1)
	}

	args := flag.Args()

	if len(args) != 1 {
		slog.Error("Usage: am2converter [INPUT]")
		os.Exit(-1)
	}

	filename := args[0]
	input, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("read file", "error", err)
		os.Exit(-2)

	}

	if !am2manager.IsAm2(input) && !am2manager.IsAm2Data(input) {
		slog.Error("invalid data type", "file", filename)
		os.Exit(-3)
	}

	var am2data am2manager.Am2Data
	if err := am2data.UnmarshalBinary(input); err != nil {
		slog.Error("unmarshal binary", "error", err)
		os.Exit(-4)
	}

	am2data.Level = byte(level)
	am2data.Mix = byte(mix)
	am2data.GainMin = byte(gainMin)
	am2data.GainMax = byte(gainMax)

	output, err := am2data.MarshalBinary()
	if err != nil {
		slog.Error("marshal binary", "error", err)
		os.Exit(-5)
	}
	fmt.Print(output)
}

func validate() error {
	if level > 255 {
		return fmt.Errorf("level (%d) must be a value between 0 and 255", level)
	}

	if mix > 100 {
		return fmt.Errorf("mix (%d) must be a value between 0 and 100", mix)
	}

	if gainMin > 100 {
		return fmt.Errorf("gain-min (%d) must be a value between 0 and 100", gainMin)
	}

	if gainMax > 100 {
		return fmt.Errorf("gain-max (%d) must be a value between 0 and 100", gainMax)
	}

	if gainMin > gainMax {
		return fmt.Errorf("gain-min (%d) cannot be greater than gain-max (%d)", gainMin, gainMax)
	}

	return nil
}
