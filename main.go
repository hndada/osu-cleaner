package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

var (
	root       string // osu! Songs dir
	cwd        string
	sameVolume bool

	banModes   = make(map[int]bool)
	banVideo   bool
	banImage   bool
	banMappers = make(map[string]bool)
	keep       = make(map[int]bool)

	noID     = -1
	blankImg = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1, 1}})
)

// dir 안 만들고 파일 이동 / 복사하면 에러 발생 추정

func main() {
	loadConfig()
	loadKeep()
	killDouble()
	sweep()
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
			fmt.Println("a")
			if sum == songSum[existName] {
				fmt.Println("c")
				os.RemoveAll(songPath)
			} else {
				fmt.Println("d")
				err = copy.Copy(songPath, filepath.Join("moved", songName))
				if err != nil {
					fmt.Println(err)
				}
			}
		} else {
			fmt.Println("b")
			marked[id] = songName
			songSum[songName] = sum
		}
	}
}

func sweep() {
	songs, err := ioutil.ReadDir(root)
	check(err)

	var info beatmapInfo
	var songName, songPath string
	var mapName, mapPath string
	var targetPath string
	isPicRemoved := make(map[string]bool)
	safe := make(map[string]bool)
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
				if err = move(mapPath); err != nil {
					fmt.Println(err)
				}
			} else {
				safe[songName] = true
			}
			if banVideo {
				targetPath = filepath.Join(songPath, info.vidName)
				if err = move(targetPath); err != nil {
					fmt.Println(err)
				}
			}
			if banImage {
				targetPath = filepath.Join(songPath, info.bgName)
				if !isPicRemoved[targetPath] {
					if err = move(targetPath); err != nil {
						fmt.Println(err)
					}
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
