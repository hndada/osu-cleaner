package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func move(srcPath string) error {
	rel, err := filepath.Rel(root, srcPath)
	if err != nil {
		return err
	}
	dest := filepath.Join(cwd, "moved", rel)
	fmt.Println(dest)
	if sameVolume {
		err := os.Rename(srcPath, dest)
		return err
	}

	fmt.Println("1")
	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	fmt.Println("2")
	out, err := os.Create(dest)
	if err != nil {
		in.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer out.Close()
	fmt.Println("3")
	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	fmt.Println("4")
	err = os.Remove(srcPath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	fmt.Println("5")
	return nil
}

func moveAll(songName string) error {
	songPath := filepath.Join(root, songName)
	err := filepath.Walk(songPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			move(path)
			return nil
		})
	check(err)
	fs, err := ioutil.ReadDir(songPath)
	check(err)
	if len(fs) != 0 {
		return errors.New("there're unmoved remained files!")
	}
	return os.Remove(songPath)
}

func blank(imgPath string) error {
	f, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	png.Encode(f, blankImg)
	return nil
}

func getMd5(mapPath string) [16]byte {
	content, err := ioutil.ReadFile(mapPath)
	check(err)
	return md5.Sum(content)
}

func addMd5(a, b [16]byte) [16]byte {
	var c [16]byte
	for i := range a {
		c[i] = a[i] + b[i]
	}
	return c
}

func getSetID(songName string) int {
	s := strings.SplitN(songName, " ", 2)[0]
	id, err := strconv.Atoi(s)
	if err != nil {
		id = noID
		noID--
	}
	return id
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// func containsInt(s []int, e int) bool {
// 	for _, v := range s {
// 		if v == e {
// 			return true
// 		}
// 	}
// 	return false
// }

// func containsStr(s []string, e string) bool {
// 	for _, v := range s {
// 		if v == e {
// 			return true
// 		}
// 	}
// 	return false
// }
