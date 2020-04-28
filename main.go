package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
)

// 블캣 다운로더에 엔터 쳐야 끝나게.

var (
	sameVolume bool
	root       string // osu! Songs dir
	banModes   = make(map[int]bool)
	banVideo   bool
	banImage   bool
	banMappers = make(map[string]bool)
	keep       = make(map[int]bool)

	noID         = -1
	safe         = make(map[string]bool)
	isPicRemoved = make(map[string]bool)

	blankImg = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1, 1}})
)

func main() {
	loadConfig()
	killDouble()
	var info beatmapInfo
	var songName, songPath string
	var mapName, mapPath string
	var targetPath string
	songs, err := ioutil.ReadDir(root)
	check(err)
	for _, song := range songs {
		if !song.IsDir() {
			continue
		}
		songName = song.Name()
		songPath = filepath.Join(root, songName)
		safe[songName] = false
		beatmaps, err := ioutil.ReadDir(songPath)
		check(err)
		for _, beatmap := range beatmaps {
			mapName = beatmap.Name()
			if filepath.Ext(mapName) != ".osu" {
				continue
			}
			mapPath = filepath.Join(songPath, mapName)
			info = getInfo(mapPath)
			if keep[info.setID] {
				safe[songName] = true
				break
			}

			if banModes[info.mode] || banMappers[info.mapper] {
				move(mapPath)
			} else {
				safe[songName] = true
			}
			if banVideo {
				targetPath = filepath.Join(songPath, info.vidName)
				move(targetPath)
			}
			if banImage {
				targetPath = filepath.Join(songPath, info.bgName)
				if !isPicRemoved[targetPath] {
					move(targetPath)
					if err = blank(targetPath); err != nil {
						fmt.Println(err)
					}
					isPicRemoved[targetPath] = true
				}
			}
		}
	}
	for _, song := range songs {
		songName = song.Name()
		if !song.IsDir() || safe[songName] {
			continue
		}
		if err = moveAll(songName); err != nil {
			fmt.Println(err)
		}
	}
}

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
	var id int
	for _, song := range songs {
		if !song.IsDir() {
			continue
		}
		var sum [16]byte
		songName = song.Name()
		id = getSetID(songName)
		songPath = filepath.Join(root, songName)
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
			} else {
				moveAll(songName)
			}
		}
	}
}
