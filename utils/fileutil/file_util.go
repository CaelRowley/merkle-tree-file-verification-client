package fileutil

import (
	"fmt"
	"log"
	"os"
	"sort"
)

type File struct {
	Name string
	Data []byte
}

func MakeDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		fmt.Println(err)
	}
}

func RemoveDir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Println(err)
	}
}

func WriteDummyFiles(path string, amount int) {
	for i := 0; i < amount; i++ {
		fileName := fmt.Sprintf("%d.txt", i)
		fileContent := fmt.Sprintf("Hello %d", i)
		WriteFile(path, fileName, fileContent)
	}
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

func GetFiles(path string) []File {
	pageSize := 1024
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	var allDirs []os.DirEntry
	for {
		dirs, err := dir.ReadDir(pageSize)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatal(err)
		}

		if len(dirs) == 0 {
			break
		}

		allDirs = append(allDirs, dirs...)
	}

	sort.Slice(allDirs, func(i int, j int) bool {
		return allDirs[i].Name() < allDirs[j].Name()
	})

	var allFiles []File
	for _, dir := range allDirs {
		file, err := GetFile(path + "/" + dir.Name())
		if err != nil {
			log.Fatal(err)
		}
		newFile := File{
			Name: dir.Name(),
			Data: file,
		}
		allFiles = append(allFiles, newFile)
	}

	return allFiles
}
