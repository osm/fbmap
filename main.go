package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type MarkerData struct {
	zone    string
	goal    string
	viewOfs string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s map_file.qc\n", os.Args[0])
		os.Exit(1)
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	result, err := convertFBMapToKTXBot(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func convertFBMapToKTXBot(qcCode string) (string, error) {
	var lines []string

	nodeRegex := regexp.MustCompile(`N\('([^']+)'\)`)
	nodeMatches := nodeRegex.FindAllStringSubmatch(qcCode, -1)
	for _, match := range nodeMatches {
		lines = append(lines, fmt.Sprintf("CreateMarker %s", match[1]))
	}

	markerData := make(map[int]MarkerData)

	zoneRegex := regexp.MustCompile(`Z(\d+)\(m(\d+)\)`)
	zoneMatches := zoneRegex.FindAllStringSubmatch(qcCode, -1)
	for _, match := range zoneMatches {
		zone := match[1]
		marker, _ := strconv.Atoi(match[2])
		data := markerData[marker]
		data.zone = zone
		markerData[marker] = data
	}

	goalRegex := regexp.MustCompile(`G(\d+)\(m(\d+)\)`)
	goalMatches := goalRegex.FindAllStringSubmatch(qcCode, -1)
	for _, match := range goalMatches {
		goal := match[1]
		marker, _ := strconv.Atoi(match[2])
		data := markerData[marker]
		data.goal = goal
		markerData[marker] = data
	}

	viewOfsRegex := regexp.MustCompile(`m(\d+)\.view_ofs_z=(\d+)`)
	viewOfsMatches := viewOfsRegex.FindAllStringSubmatch(qcCode, -1)
	for _, match := range viewOfsMatches {
		marker, _ := strconv.Atoi(match[1])
		offset := match[2]
		data := markerData[marker]
		data.viewOfs = offset
		markerData[marker] = data
	}

	paths := make(map[int]map[int]string)
	pathRegex := regexp.MustCompile(`m(\d+)\.P(\d+)=m(\d+)`)
	pathMatches := pathRegex.FindAllStringSubmatch(qcCode, -1)
	for _, match := range pathMatches {
		marker, _ := strconv.Atoi(match[1])
		pathIdx, _ := strconv.Atoi(match[2])
		target := match[3]
		if paths[marker] == nil {
			paths[marker] = make(map[int]string)
		}
		paths[marker][pathIdx] = target
	}

	flags := make(map[int]map[int]string)
	flagRegex := regexp.MustCompile(`m(\d+)\.D(\d+)=(\d+)`)
	flagMatches := flagRegex.FindAllStringSubmatch(qcCode, -1)
	for _, match := range flagMatches {
		marker, _ := strconv.Atoi(match[1])
		flagIdx, _ := strconv.Atoi(match[2])
		value := match[3]
		if flags[marker] == nil {
			flags[marker] = make(map[int]string)
		}
		flags[marker][flagIdx] = value
	}

	markerNums := make([]int, 0, len(markerData))
	for k := range markerData {
		markerNums = append(markerNums, k)
	}
	sort.Ints(markerNums)

	for _, marker := range markerNums {
		data := markerData[marker]
		if data.goal != "" {
			lines = append(lines, fmt.Sprintf("SetGoal %d %s", marker, data.goal))
		}
		if data.zone != "" {
			lines = append(lines, fmt.Sprintf("SetZone %d %s", marker, data.zone))
		}
		if data.viewOfs != "" {
			lines = append(lines,
				fmt.Sprintf("SetMarkerViewOfs %d %s", marker, data.viewOfs))
		}
	}

	pathMarkers := make([]int, 0, len(paths))
	for k := range paths {
		pathMarkers = append(pathMarkers, k)
	}
	sort.Ints(pathMarkers)

	for _, marker := range pathMarkers {
		markerPaths := paths[marker]
		pathIndices := make([]int, 0, len(markerPaths))
		for k := range markerPaths {
			pathIndices = append(pathIndices, k)
		}
		sort.Ints(pathIndices)

		for _, idx := range pathIndices {
			lines = append(lines,
				fmt.Sprintf("SetMarkerPath %d %d %s", marker, idx, markerPaths[idx]))
		}
	}

	flagMarkers := make([]int, 0, len(flags))
	for k := range flags {
		flagMarkers = append(flagMarkers, k)
	}
	sort.Ints(flagMarkers)

	for _, marker := range flagMarkers {
		markerFlags := flags[marker]
		flagIndices := make([]int, 0, len(markerFlags))
		for k := range markerFlags {
			flagIndices = append(flagIndices, k)
		}
		sort.Ints(flagIndices)

		for _, idx := range flagIndices {
			value := markerFlags[idx]
			var flag string
			switch value {
			case "512":
				flag = "j"
			case "1024":
				flag = "r"
			}
			if flag != "" {
				lines = append(lines,
					fmt.Sprintf("SetMarkerPathFlags %d %d %s", marker, idx, flag))
			}
		}
	}

	return strings.Join(lines, "\n"), nil
}
