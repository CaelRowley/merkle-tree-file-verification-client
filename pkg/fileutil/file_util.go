package fileutil

import (
	"fmt"
	"os"
	"sync"
)

type File struct {
	Name string
	Data []byte
}

func MakeDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Println(err)
	}
}

func RemoveDir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Println(err)
	}
}

func WriteDummyFiles(path string, amount int, ch chan<- int) {
	maxGoroutines := 4
	sem := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup
	wg.Add(amount)

	for i := 0; i < amount; i++ {
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			fileName := fmt.Sprintf("%d.txt", i)
			fileContent := fmt.Sprintf("Hello %d", i)
			WriteFile(path, fileName, fileContent)
			ch <- i
		}(i)
	}

	wg.Wait()
}

func WriteFile(path string, name string, content string) {
	file, err := os.Create(path + "/" + name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GetFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetFiles(path string, ch chan<- int) []File {
	pageSize := 1024
	dir, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer dir.Close()

	var allDirs []os.DirEntry
	for {
		dirs, err := dir.ReadDir(pageSize)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
		}
		if len(dirs) == 0 {
			break
		}
		allDirs = append(allDirs, dirs...)
	}

	var allFiles []File
	for i, dir := range allDirs {
		file, err := GetFile(path + "/" + dir.Name())
		if err != nil {
			fmt.Println(err)
		}
		newFile := File{
			Name: dir.Name(),
			Data: file,
		}
		allFiles = append(allFiles, newFile)
		ch <- i
	}

	return allFiles
}
