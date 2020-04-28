// no longer use

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getSetID(songName string) int {
	s := strings.SplitN(songName, " ", 2)[0]
	id, err := strconv.Atoi(s)
	if err != nil {
		// fmt.Printf("%s gets temp id %d.\n", songName, id)
		fmt.Printf("%s has no id.\n", songName)
		id = -1
	}
	return id
}

func isModeMatch(path string) (bool, error) {

	var s string
	scanner := bufio.NewScanner(f) // Splits on newlines by default.
	for scanner.Scan() {
		s = scanner.Text()
		if strings.HasPrefix(s, "Mode: ") {
			if contains(banModes, strings.Split(s, ": ")[1]) {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return false, scanner.Err()
}

func getBgName(mapPath string) string {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	var s string
	var event bool
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s = scanner.Text()
		if s == "[Event]" {
			event = true
		}
		if strings.HasPrefix(s, "0,0,") && event {
			info.bgName = strings.Trim(strings.Split(s, ",")[2], "\"")
		}
	}
	return ""
}