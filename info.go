package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type beatmapInfo struct {
	mode     int
	metadata map[string]string
	mapID    int
	setID    int
	mapper   string
	bgName   string
	vidName  string
}

func getInfo(mapPath string) beatmapInfo {
	f, err := os.Open(mapPath)
	if err != nil {
		fmt.Printf("%s: failed to load file\n", mapPath)
		return beatmapInfo{}
	}
	defer f.Close()

	var line, section string
	var splitKeyValue []string
	var mode int
	var info beatmapInfo
	metadata := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		switch {
		case line == "":
			break
		case isSection(line):
			section = strings.Trim(line, "[]")
		default:
			switch section {
			case "General":
				if strings.HasPrefix(line, "Mode: ") {
					mode, err = strconv.Atoi(strings.Split(line, ": ")[1])
					if err != nil {
						fmt.Printf("%s: failed to load mode\n", mapPath)
						return info
					}
					info.mode = mode
				}
			case "Metadata":
				splitKeyValue = strings.Split(line, `:`)
				metadata[splitKeyValue[0]] = splitKeyValue[1]
			case "Events":
				if strings.HasPrefix(line, "0,0,") {
					info.bgName = strings.Trim(strings.Split(line, ",")[2], "\"")
				}
				if strings.HasPrefix(line, "Video,") || strings.HasPrefix(line, "1,") {
					info.vidName = strings.Trim(strings.Split(line, ",")[2], "\"")
				}
			}
		}
	}
	info.mapID, info.setID = getID(metadata)
	info.mapper = getMapper(metadata)
	info.metadata = metadata
	return info
}

func getMapper(metadata map[string]string) string {
	values := strings.SplitN(metadata["Version"], "'s", 2)
	if len(values) == 1 {
		return metadata["Creator"]
	}
	return values[0]
}

func getID(metadata map[string]string) (int, int) {
	mapID, err := strconv.Atoi(metadata["BeatmapID"])
	if err != nil {
		mapID = -1
	}
	setID, err := strconv.Atoi(metadata["BeatmapSetID"])
	if err != nil {
		setID = -1
	}
	return mapID, setID
}

func isSection(line string) bool {
	if len(line) == 0 {
		return false
	}
	return string(line[0]) == "[" && string(line[len(line)-1]) == "]"
}
