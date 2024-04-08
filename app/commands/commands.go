package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/api"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/utils/merkletree"
)

const (
	TestFilePath     = "files/dummy"
	DownloadFilePath = "files/downloads"
	CorruptFilePath  = "files/corrupt.txt"
)

func CreateFilesCmd() {
	fileutil.MakeDir(TestFilePath)
	prompt := promptui.Prompt{
		Label: "Amount to create",
	}
	input, err := prompt.Run()
	if err != nil {
		fmt.Println("Error with prompt:", err)
		return
	}
	amount, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Please enter an integer:", err)
		return
	}

	ch := startLoading("Deleting previous test files")
	start := time.Now()
	deleteFilesInDir(TestFilePath)
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Test files deleted! %s\n", elapsed)

	chLoading, chCount := startLoadingWithCount("Creating %d/%d test files", amount)
	start = time.Now()
	fileutil.WriteDummyFiles(TestFilePath, amount, chCount)
	elapsed = time.Since(start)
	endLoadingWithCount(chLoading, chCount)

	cwd, _ := os.Getwd()
	fmt.Printf("%d test files created in:\n%s/%s %s\n", amount, cwd, TestFilePath, elapsed)
}

func CreateTreeCmd() {
	fileutil.MakeDir(TestFilePath)

	chLoading, chCount := startLoadingWithCount("Reading %d files", 0)
	start := time.Now()
	files := fileutil.GetFiles(TestFilePath, chCount)
	elapsed := time.Since(start)
	endLoadingWithCount(chLoading, chCount)
	fmt.Printf("Files read %s\n", elapsed)

	if len(files) < 1 {
		fmt.Println("Please create some test files first.")
		return
	}

	chLoading = startLoading("Hashing files")
	start = time.Now()
	var fileHashes [][]byte
	for _, file := range files {
		fileHash := sha256.Sum256([]byte(file.Data))
		fileHashes = append(fileHashes, fileHash[:])
	}
	sort.Slice(fileHashes, func(i int, j int) bool {
		return bytes.Compare(fileHashes[i], fileHashes[j]) < 0
	})
	elapsed = time.Since(start)
	endLoading(chLoading)
	fmt.Printf("Files hashed %s\n", elapsed)

	chLoading = startLoading("Building tree")
	start = time.Now()
	merkletree.BuildTree(fileHashes)
	rootHash := hex.EncodeToString(merkletree.Root.Hash[:])
	elapsed = time.Since(start)
	endLoading(chLoading)
	fmt.Printf("Generated Merkle tree %s\n", elapsed)

	fmt.Printf("Root hash: %s\n", rootHash)
}

func UploadFilesCmd(serverURL string) {
	fileutil.MakeDir(TestFilePath)

	chLoading, chCount := startLoadingWithCount("Reading %d files", 0)
	start := time.Now()
	files := fileutil.GetFiles(TestFilePath, chCount)
	elapsed := time.Since(start)
	endLoadingWithCount(chLoading, chCount)
	fmt.Printf("%d test files read %s\n", len(files), elapsed)

	if len(files) < 1 {
		fmt.Println("Please create some test files first.")
		return
	}

	err := api.DeleteAllFiles(serverURL)
	if err != nil {
		fmt.Println("Error deleting files in the DB:", err)
		return
	}
	fmt.Println("Deleted all files in the DB!")

	chLoading, chCount = startLoadingWithCount("Uploading %d files", 0)
	start = time.Now()
	err = api.UploadFiles(serverURL, files, chCount)
	elapsed = time.Since(start)
	endLoadingWithCount(chLoading, chCount)
	if err != nil {
		fmt.Println("Error sending the files to the server:", err)
		return
	}

	fmt.Printf("Uploaded %d files! %s\n", len(files), elapsed)
	fmt.Printf("IDs range from 1 to %d\n", len(files))
}

func DownloadAndVerifyFileCmd(serverURL string) {
	fileutil.MakeDir(DownloadFilePath)

	if merkletree.Root == nil {
		fmt.Println("You need to Generate a Merkle tree first.")
		return
	}

	prompt := promptui.Prompt{
		Label: "Enter file id",
	}
	input, err := prompt.Run()
	if err != nil {
		fmt.Println("Error with prompt:", err)
		return
	}
	_, err = strconv.Atoi(input)
	if err != nil {
		fmt.Println("Please enter an integer:", err)
		return
	}

	start := time.Now()
	fileName, fileData, err := api.GetFile(serverURL, input)
	if err != nil {
		fmt.Println("Error getting file with id:", input, ":", err)
		return
	}

	filePath := DownloadFilePath + "/" + fileName
	err = os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		fmt.Println("Error getting file with id:", input, ":", err)
		return
	}

	proof, err := api.GetProof(serverURL, input)
	if err != nil {
		fmt.Println("Error getting proof:", err)
		return
	}

	elapsed := time.Since(start)
	cwd, _ := os.Getwd()
	fmt.Printf("Downloaded file %s and proof to:\n%s/%s %s\n", input, cwd, filePath, elapsed)

	fileHash := sha256.Sum256(fileData)

	start = time.Now()
	isVerified, proofRoot := merkletree.VerifyMerkleProof(merkletree.Root.Hash, fileHash[:], proof)
	rootHash := hex.EncodeToString(merkletree.Root.Hash)
	proofRootHash := hex.EncodeToString(proofRoot)
	elapsed = time.Since(start)
	fmt.Println("New root generated with Merkle proof!", elapsed)
	fmt.Printf("Stored root hash: %s\n", rootHash)
	fmt.Printf("Proof root hash:  %s\n", proofRootHash)
	if isVerified {
		fmt.Printf("The hashes match!\n%s has not been modified\n", fileName)
	} else {
		fmt.Printf("The hashes don't match!\n%s has been corrupted\n", fileName)
	}
}

func CorruptFileCmd(serverURL string) {
	prompt := promptui.Prompt{
		Label: "Enter file id",
	}
	input, err := prompt.Run()
	if err != nil {
		fmt.Println("Error with prompt:", err)
		return
	}
	_, err = strconv.Atoi(input)
	if err != nil {
		fmt.Println("Please enter an integer:", err)
		return
	}

	start := time.Now()
	file, err := fileutil.GetFile(CorruptFilePath)
	if err != nil {
		fmt.Println("Error getting corrupt file:", err)
		return
	}

	err = api.CorruptFile(serverURL, input, file)
	if err != nil {
		fmt.Println("Error corrupting file in DB:", err)
		return
	}
	elapsed := time.Since(start)

	fmt.Printf("File %s has been modified on the server! %s\n", input, elapsed)
}

func DeleteTestFilesCmd() {
	ch := startLoading("Deleting test files")
	start := time.Now()
	deleteFilesInDir(TestFilePath)
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Test files deleted! %s\n", elapsed)
}

func DeleteDownloadsCmd() {
	ch := startLoading("Deleting downloads")
	start := time.Now()
	deleteFilesInDir(DownloadFilePath)
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Downloads deleted! %s\n", elapsed)
}

func ExitCmd() {
	fmt.Println("Au revoir!")
	os.Exit(0)
}

func deleteFilesInDir(path string) {
	fileutil.RemoveDir(path)
	fileutil.MakeDir(path)
}

func startLoading(text string) chan bool {
	ch := make(chan bool)
	dots := []string{text + "   ", text + ".  ", text + ".. ", text + "..."}

	go func() {
		for {
			for _, dot := range dots {
				select {
				case <-ch:
					fmt.Print("\r\033[K")
					return
				default:
					fmt.Printf("\r%s", dot)
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
	}()

	return ch
}

func startLoadingWithCount(text string, total int) (chan bool, chan int) {
	ch := make(chan bool)
	chCount := make(chan int)
	count := 0
	message := ""
	if total > 0 {
		message = fmt.Sprintf(text, count, total)
	} else {
		message = fmt.Sprintf(text, count)
	}
	dots := []string{message + "   ", message + ".  ", message + ".. ", message + "..."}

	go func() {
		for {
			select {
			case <-ch:
				fmt.Print("\r\033[K")
				return
			default:
				count = <-chCount
				if total > 0 {
					message = fmt.Sprintf(text, count, total)
				} else {
					message = fmt.Sprintf(text, count)
				}
				dots = []string{message + "   ", message + ".  ", message + ".. ", message + "..."}
			}
		}
	}()

	go func() {
		for {
			for _, dot := range dots {
				select {
				case <-ch:
					fmt.Print("\r\033[K")
					return
				default:
					fmt.Printf("\r%s", dot)
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
	}()

	return ch, chCount
}

func endLoading(ch chan bool) {
	close(ch)
	fmt.Print("\r\033[K")
}

func endLoadingWithCount(ch chan bool, chFiles chan int) {
	close(ch)
	close(chFiles)
	fmt.Print("\r\033[K")
}
