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
	if strings.Contains(relPath, ":\\") { // old maps are vulnerable to this
		return nil
	}
	destPath := filepath.Join(cwd, "moved", relPath)
	err := os.MkdirAll(filepath.Dir(destPath), 0777)
	check(err)

	absPath := filepath.Join(root, relPath)
	in, err := os.Open(absPath)
	if err != nil {
		return err
	}

	if sameVolume {
		in.Close()
		err = os.Rename(absPath, destPath)
		return err
	}

	out, err := os.Create(destPath)
	if err != nil {
		in.Close()
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return err
	}

	err = os.Remove(absPath)
	if err != nil {
		return err
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

func dirSize(path string) int64 {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		fmt.Println(err)
	}
	return size
}

func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
