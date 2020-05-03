package main

import (
	"image"
	"io/ioutil"
	"os"
	"path/filepath"

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

	noID     = -1
	blankImg = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1, 1}})
)

func main() {
	loadConfig()
	loadKeep()
	killDouble()
	sweep()
}

// check sameness with sums of all beatmap's md5 of each folder
func killDouble() {
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
				os.RemoveAll(songPath)
			} else { // remove older and update with newer
				older, newer = olderNewer(filepath.Join(root, existName), songPath)
				f, err := os.Create(filepath.Join("doubled", filepath.Base(older)+".zip"))
				check(err)
				err = zip.Archive(older, f, nil)
				check(err)
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

	var info beatmapInfo
	var songName, songPath string
	var mapName, mapPath string
	movesList := make(map[string]bool)
	allBan := make(map[string]bool)
	bgPathsList := make(map[string]bool)
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
				allBan[songName] = false
				break
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
				info = getInfo(osbName)
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
		os.RemoveAll(songPath)
	}

	for relPath := range movesList {
		err = move(relPath)
		check(err)
	}
	// put blank pics; osu! gets angry when it detects bg was deleted
	for relPath := range bgPathsList {
		err = blank(relPath)
		check(err)
	}
}
