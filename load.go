package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	ledDriverFile = "/proc/acpi/nuc_led"
	loadAvgFile = "/proc/loadavg"
	loadFeedInterval = 5 * time.Second
)

type ledColorConfig struct {
	load float64
	color string
	brightness int
	blink string
}

var colorsByLoad = []ledColorConfig {
	{0, "white", 80, "none"},
	{0.02, "blue", 10, "none"},
	{0.05, "blue", 20, "none"},
	{0.10, "blue", 40, "none"},
	{0.15, "blue", 80, "none"},
	{0.25, "cyan", 50, "none"},
	{0.5, "green", 80, "none"},
	{0.75, "yellow", 60, "none"},
	{1.0, "pink", 60, "none"},
	{2.0, "red", 80, "none"},
	{4.0, "red", 100, "none"},
	{0, "red", 100, "fade_fast"},
}

func handleShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log("shutting down", "|", "signal:", sig)
		setRingColor(ledColorConfig{0, "off", 0, "none"})
		os.Exit(0)
	}()
}

func main() {
	handleShutdown()
	cores := float64(runtime.NumCPU())
	for i, _ := range colorsByLoad {
		colorsByLoad[i].load *= cores
	}
	log("starting", "|", "cores:", cores, "|", "colors:", colorsByLoad)
	loadAverageMonitor()
}

func getColorByLoad(load float64) ledColorConfig {
	for _, i := range colorsByLoad {
		if load < i.load {
			return i
		}
	}
	return colorsByLoad[len(colorsByLoad) - 1]
}

func setRingColor(color ledColorConfig) {
	cmd := fmt.Sprintf("ring,%d,%s,%s", color.brightness, color.blink, color.color)
	err := ioutil.WriteFile(ledDriverFile, []byte(cmd), 0644)
	if err != nil {
		log("setRingColor error:", err)
	}
}

func getLoadAverage() (float64, error) {
	load, err := ioutil.ReadFile(loadAvgFile)
	if err != nil {
		return -1, err
	}
	ls := strings.Split(string(load), " ")
	if len(ls) < 1 {
		return -1, errors.New("no data")
	}
	lv, err := strconv.ParseFloat(ls[0], 64)
	if err != nil {
		return -1, err
	}
	return lv, nil
}

func loadAverageMonitor() {
	var prevColor ledColorConfig
	for {
		load, err := getLoadAverage()
		if err != nil {
			log("getLoadAverage error:", err)
		} else {
			var color = getColorByLoad(load)
			if color != prevColor {
				log("load:", load, "|", "color:", color)
				setRingColor(color)
				prevColor = color
			}
		}
		time.Sleep(loadFeedInterval)
	}
}

func log(message ...interface{}) {
	prefix := []interface{}{time.Now().Format(time.RFC1123), "|"}
	fmt.Println(append(prefix, message...)...)
}
