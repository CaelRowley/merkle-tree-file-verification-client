package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/api"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/utils/merkletree"
)

const (
	TEST_FILE_PATH     = "files/test"
	DOWNLOAD_FILE_PATH = "files/downloads"
	CORRUPT_FILE_PATH  = "files/corrupt.txt"
)

func CreateFilesCmd() {
	chLoading := startLoading("Deleting previous test files")
	deleteFilesInDir(TEST_FILE_PATH)
	endLoading(chLoading)

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

	chLoading = startLoading(fmt.Sprintf("Creating %d test files", amount))
	start := time.Now()
	fileutil.WriteDummyFiles(TEST_FILE_PATH, amount)
	elapsed := time.Since(start)
	endLoading(chLoading)

	fmt.Printf("%d test files created %s\n\n", amount, elapsed)
}

func CreateTreeCmd() {
	start := time.Now()
	chLoading := startLoading("Reading files and hashing data")
	files := fileutil.GetFiles(TEST_FILE_PATH)
	if len(files) < 1 {
		endLoading(chLoading)
		fmt.Println("Please create some test files first.")
		return
	}

	var fileHashes [][]byte
	for _, file := range files {
		fileHash := sha256.Sum256([]byte(file.Data))
		fileHashes = append(fileHashes, fileHash[:])
	}
	elapsed := time.Since(start)
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
	chLoading := startLoading("Reading test files")
	start := time.Now()
	files := fileutil.GetFiles(TEST_FILE_PATH)
	elapsed := time.Since(start)
	endLoading(chLoading)

	if len(files) < 1 {
		fmt.Println("Please create some test files first.")
		return
	}

	fmt.Printf("%d test files read %s\n", len(files), elapsed)

	err := api.DeleteAllFiles(serverURL)
	if err != nil {
		fmt.Println("Error deleting files in the DB:", err)
		return
	}
	fmt.Println("Deleted all files in the DB!")

	chLoading = startLoading(fmt.Sprintf("Uploading %d files", len(files)))
	start = time.Now()
	err = api.UploadFiles(serverURL, files)
	endLoading(chLoading)
	if err != nil {
		fmt.Println("Error sending the files to the server:", err)
		return
	}
	elapsed = time.Since(start)
	fmt.Printf("Uploaded %d files! %s\n\n", len(files), elapsed)
}

func DownloadAndVerifyFileCmd(serverURL string) {
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

	start := time.Now()
	fileName, fileData, err := api.GetFile(serverURL, input)
	if err != nil {
		fmt.Println("Error getting file with id:", input, ":", err)
		return
	}

	filePath := DOWNLOAD_FILE_PATH + "/" + fileName
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

	fileHash := sha256.Sum256(fileData)
	elapsed := time.Since(start)
	fmt.Printf("Downloaded file %s to: %s %s\n", input, filePath, elapsed)

	start = time.Now()
	isVerified, proofRoot := merkletree.VerifyMerkleProof(merkletree.Root.Hash, fileHash[:], proof)
	rootHash := hex.EncodeToString(merkletree.Root.Hash)
	proofRootHash := hex.EncodeToString(proofRoot)
	elapsed = time.Since(start)
	fmt.Println("New root generated with Merkle proof!", elapsed)
	fmt.Printf("Stored root hash: %s\n", rootHash)
	fmt.Printf("Proof root hash:  %s\n", proofRootHash)
	if isVerified {
		fmt.Printf("The hashes match!\n%s has not been modified\n", filePath)
	} else {
		fmt.Printf("The hashes don't match!\n%s has been corrupted\n", filePath)
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

	start := time.Now()
	file, err := fileutil.GetFile(CORRUPT_FILE_PATH)
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
	fmt.Printf("File %s has been modified on the server! %s\n\n", input, elapsed)
}

func DeleteTestFilesCmd() {
	ch := startLoading("Deleting test files")
	start := time.Now()
	deleteFilesInDir(TEST_FILE_PATH)
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Test files deleted! %s\n\n", elapsed)
}

func DeleteDownloadsCmd() {
	start := time.Now()
	ch := startLoading("Deleting downloads")
	deleteFilesInDir(DOWNLOAD_FILE_PATH)
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Downloads deleted! %s\n\n", elapsed)
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

func endLoading(ch chan bool) {
	close(ch)
	fmt.Print("\r\033[K")
}
