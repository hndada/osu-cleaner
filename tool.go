package main

import (
	"crypto/md5"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// input path must be filepath
func move(relPath string) error {
	destPath := filepath.Join(cwd, "moved", relPath)
	err := os.MkdirAll(filepath.Dir(destPath), 0777)
	check(err)
	// if s, _:= os.Stat(destPath); s.IsDir()

	absPath := filepath.Join(root, relPath)
	if sameVolume {
		err = os.Rename(absPath, destPath)
		return err
	}

	in, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		in.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}

	err = os.Remove(absPath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}

func blank(imgRelPath string) error {
	f, err := os.Create(filepath.Join(root, imgRelPath))
	if err != nil {
		return err
	}
	png.Encode(f, blankImg)
	return nil
}

func olderNewer(path1, path2 string) (string, string) {
	f1, err := os.Stat(path1)
	check(err)
	f2, err := os.Stat(path2)
	check(err)
	if f1.ModTime().Before(f2.ModTime()) {
		return path1, path2
	}
	return path2, path1
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
		panic(err)
	}
}

// func moveAll(songName string) error {
// 	songPath := filepath.Join(root, songName)
// 	err := filepath.Walk(songPath,
// 		func(path string, info os.FileInfo, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			move(path)
// 			return nil
// 		})
// 	check(err)
// 	fs, err := ioutil.ReadDir(songPath)
// 	check(err)
// 	if len(fs) != 0 {
// 		return errors.New("there're unmoved remained files!")
// 	}
// 	return os.Remove(songPath)
// }

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
