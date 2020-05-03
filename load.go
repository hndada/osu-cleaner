package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func loadConfig() {
	f, err := os.Open("config.txt")
	check(err)
	defer f.Close()

	var text string
	var values []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text = scanner.Text()
		if text == "" || strings.HasPrefix(text, "//") {
			continue
		}
		values = strings.SplitN(text, ":", 2)
		if len(values) < 2 {
			continue
		}
		switch values[0] {
		case "Songs":
			root = values[1]
			if _, err := os.Stat(root); err != nil {
				log.Fatalf("invalid: %s is not a valid directory.", root)
			}
		case "BanModes":
			modes := strings.Split(values[1], ",")
			for _, m := range modes {
				v, err := strconv.Atoi(m)
				check(err)
				banModes[v] = true
			}
		case "BanVideo":
			switch values[1] {
			case "1":
				banVideo = true
			case "0":
				banVideo = false
			default:
				log.Fatal("invalid: value should be 1 or 0.")
			}
		case "BanImage":
			switch values[1] {
			case "1":
				banImage = true
			case "0":
				banImage = false
			default:
				log.Fatal("invalid: value should be 1 or 0.")
			}
		case "BanStoryboard":
			switch values[1] {
			case "1":
				banSB = true
			case "0":
				banSB = false
			default:
				log.Fatal("invalid: value should be 1 or 0.")
			}
		}
	}
	check(scanner.Err())

	exePath, err := os.Executable()
	check(err)
	cwd = filepath.Dir(exePath)
	sameVolume = filepath.VolumeName(exePath) == filepath.VolumeName(root)
}

func loadKeep() {
	f, err := os.Open("keep.txt")
	check(err)
	defer f.Close()

	var text string
	var id int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text = scanner.Text()
		if text == "" || strings.HasPrefix(text, "//") {
			continue
		}
		id, err = strconv.Atoi(text)
		if err != nil {
			continue
		}
		keep[id] = true
	}
}

func loadBanMappers() {
	f, err := os.Open("banMapper.txt")
	check(err)
	defer f.Close()

	var text string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text = scanner.Text()
		if text == "" || strings.HasPrefix(text, "//") {
			continue
		}
		banMappers[text] = true
	}
}
