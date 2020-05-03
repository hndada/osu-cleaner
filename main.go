package main

import (
	"bufio"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pierrre/archivefile/zip"
)

var (
	root       string // osu! Songs dir
	cwd        string // current working directory
	sameVolume bool   // whether Songs dir and cleaner are in same drive volume

	banModes   = make(map[int]bool)
	banVideo   bool
	banImage   bool
	banSB      bool
	banMappers = make(map[string]bool)
	keep       = make(map[int]bool)

	size     = make(map[string]int64)
	noID     = -1
	blankImg = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1, 1}})
)

func main() {
	loadConfig()
	fmt.Printf("osu! Songs folder: %s\n", root)
	modes := make([]string, 0, len(banModes))
	for mode := range banModes {
		modes = append(modes, map[int]string{0: "Standard", 1: "Taiko", 2: "Catch", 3: "Mania"}[mode])
	}
	fmt.Printf("Banned modes: %s\n", strings.Join(modes, ","))
	fmt.Printf("Ban videos: %t\n", banVideo)
	fmt.Printf("Ban background images: %t\n", banImage)
	fmt.Printf("Ban storyboards: %t\n", banSB)
	mappers := make([]string, 0, len(banMappers))
	for mapper := range banMappers {
		mappers = append(mappers, mapper)
	}
	fmt.Printf("Banned Mappers: %s\n", strings.Join(mappers, ","))

	time.Sleep(time.Second)
	fmt.Print("\nStart? This might take several minutes to an hour. (y/n) ")
	var yes string
	_, err := fmt.Scan(&yes)
	check(err)
	yes = strings.TrimSpace(yes)
	yes = strings.ToLower(yes)
	if yes != "y" {
		os.Exit(1)
	}

	size["Songs"] = dirSize(root)
	loadKeep()
	loadBanMappers()
	fmt.Println("Loading config done, start killing doubled files...")
	killDouble()
	fmt.Println("Double killing done, start cleaning...")
	sweep()
	printSize()
	buf := bufio.NewReader(os.Stdin)
	buf.Discard(1) // I guess it's due to variable yes
	fmt.Print("\nPress the Enter to terminate the console.")
	buf.ReadBytes('\n') // wait for Enter Key
}

// check sameness with sums of all beatmap's md5 of each folder
func killDouble() {
	size["doubled_deleted"] = 0
	size["doubled_moved"] = 0
	marked := make(map[int]string)
	songSum := make(map[string][16]byte)
	if _, err := os.Stat("doubled"); os.IsNotExist(err) {
		os.Mkdir("doubled", os.ModePerm)
	}
	songs, err := ioutil.ReadDir(root)
	check(err)

	var songName, songPath string
	var mapName, mapPath string
	var older, newer string
	var id int
	for _, song := range songs {
		if !song.IsDir() {
			continue
		}
		songName = song.Name()
		id = getSetID(songName)
		songPath = filepath.Join(root, songName)

		var sum [16]byte
		beatmaps, err := ioutil.ReadDir(songPath)
		check(err)
		for _, beatmap := range beatmaps {
			mapName = beatmap.Name()
			if filepath.Ext(mapName) != ".osu" {
				continue
			}
			mapPath = filepath.Join(songPath, mapName)
			sum = addMd5(sum, getMd5(mapPath))
		}

		if existName, ok := marked[id]; ok {
			if sum == songSum[existName] {
				size["doubled_deleted"] += dirSize(songPath)
				os.RemoveAll(songPath)
			} else { // remove older and update with newer
				older, newer = olderNewer(filepath.Join(root, existName), songPath)
				f, err := os.Create(filepath.Join("doubled", filepath.Base(older)+".zip"))
				check(err)
				err = zip.Archive(older, f, nil)
				check(err)
				size["doubled_moved"] += dirSize(older)
				os.RemoveAll(older)
				marked[id] = filepath.Base(newer)
				songSum[filepath.Base(newer)] = sum
			}
		} else {
			marked[id] = songName
			songSum[songName] = sum
		}
	}
}

func sweep() {
	songs, err := ioutil.ReadDir(root)
	check(err)
	if _, err := os.Stat("moved"); os.IsNotExist(err) {
		os.Mkdir("moved", os.ModePerm)
	}

	size["moved"] = 0
	var info beatmapInfo
	var songName, songPath string
	var mapName, mapPath string
	var values []string
	var setID int
	movesList := make(map[string]bool)
	allBan := make(map[string]bool)
	bgPathsList := make(map[string]bool)
song:
	for _, song := range songs {
		var safe bool
		var osbName string
		if !song.IsDir() {
			continue
		}
		songName = song.Name()
		songPath = filepath.Join(root, songName)
		beatmaps, err := ioutil.ReadDir(songPath)
		check(err)
		moves := make([]string, 0, len(beatmaps))
		bgPaths := make([]string, 0, len(beatmaps))

		values = strings.SplitN(songName, " ", 2)
		if setID, err = strconv.Atoi(values[0]); err == nil {
			if keep[setID] {
				fmt.Printf("Song %s is kept from cleaner\n", songName)
				// safe = true
				continue
			}
		}

		for _, beatmap := range beatmaps {
			mapName = beatmap.Name()
			if banSB && filepath.Ext(mapName) == ".osb" {
				osbName = mapName
			}
			if filepath.Ext(mapName) != ".osu" {
				continue
			}
			mapPath = filepath.Join(songPath, mapName)
			info = getInfo(mapPath)
			if keep[info.setID] {
				fmt.Printf("Song %s (SetID: %d) is kept from cleaner\n", info.metadata["Title"], info.setID)
				// safe = true
				continue song
			}
			if banModes[info.mode] || banMappers[info.mapper] {
				moves = append(moves, filepath.Join(songName, mapName))
			} else {
				safe = true
			}
			if banVideo && info.vidName != "" {
				moves = append(moves, filepath.Join(songName, info.vidName))
			}
			if banImage && info.bgName != "" {
				moves = append(moves, filepath.Join(songName, info.bgName))
				bgPaths = append(bgPaths, filepath.Join(songName, info.bgName))
			}
			if banSB {
				for _, relPath := range info.sbRelPaths {
					moves = append(moves, filepath.Join(songName, relPath))
				}
			}
		}

		if !safe { // marked with allBan beatmapSet goes archieved with .osz
			allBan[songName] = true
		} else {
			if osbName != "" {
				info = getInfo(filepath.Join(songPath, osbName))
				for _, relPath := range info.sbRelPaths {
					moves = append(moves, filepath.Join(songName, relPath))
				}
			}
			for _, relPath := range moves {
				movesList[relPath] = true
			}
			for _, relPath := range bgPaths {
				bgPathsList[relPath] = true
			}
		}
	}

	for songName := range allBan {
		songPath = filepath.Join(root, songName) + "/"
		f, err := os.Create(filepath.Join(cwd, "moved", songName+".osz"))
		check(err)
		err = zip.Archive(songPath, f, nil)
		check(err)
		size["moved"] += dirSize(songPath)
		os.RemoveAll(songPath)
	}

	for relPath := range movesList {
		err = move(relPath)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	// put blank pics; osu! gets angry when it detects bg was deleted
	for relPath := range bgPathsList {
		err = blank(relPath)
		check(err)
	}
}

func printSize() {
	reduced := size["moved"] + size["doubled_moved"] + size["doubled_deleted"]
	fmt.Printf("Original Songs size: %s\n", byteCountIEC(size["Songs"]))
	fmt.Printf("After Songs size: %s\n", byteCountIEC(size["Songs"]-reduced))
	fmt.Printf("Total reduced size: %s (deleted: %s)\n",
		byteCountIEC(reduced), byteCountIEC(size["doubled_deleted"]))
}
