package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/api"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/merkletree"
)

const TEST_FILE_PATH = "testfiles"
const DOWNLOAD_FILE_PATH = "downloads"

var (
	serverURL = os.Getenv("SERVER_URL")
)

var root *merkletree.Node

func main() {
	fileutil.MakeDir(TEST_FILE_PATH)
	fileutil.MakeDir(DOWNLOAD_FILE_PATH)
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	const CREATE_FILES_CMD = "Create Test Files"
	const CREATE_TREE_COMMAND = "Generate Merkle Tree"
	const UPLOAD_FILES_CMD = "Upload Test Files"
	const DELETE_TEST_FILES_CMD = "Delete Test Files"
	const DELETE_DOWNLOADS_CMD = "Delete Downloads"
	const DOWNLOAD_AND_VERIFY_FILE_CMD = "Download and Verify File"
	const CORRUPT_FILE_CMD = "Corrupt a File on Server"
	const EXIT_CMD = "Exit"

	commands := []string{
		CREATE_FILES_CMD,
		CREATE_TREE_COMMAND,
		UPLOAD_FILES_CMD,
		DOWNLOAD_AND_VERIFY_FILE_CMD,
		CORRUPT_FILE_CMD,
		DELETE_TEST_FILES_CMD,
		DELETE_DOWNLOADS_CMD,
		EXIT_CMD,
	}

	prompt := promptui.Select{
		Label: "Select a command",
		Items: commands,
		Size:  len(commands),
	}

	for {
		_, selected, err := prompt.Run()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println()

		switch selected {
		case CREATE_FILES_CMD:
			createFilesCmd()
		case CREATE_TREE_COMMAND:
			createTreeCmd()
		case UPLOAD_FILES_CMD:
			uploadFilesCmd()
		case DELETE_TEST_FILES_CMD:
			deleteTestFilesCmd()
		case DELETE_DOWNLOADS_CMD:
			deleteDownloadsCmd()
		case DOWNLOAD_AND_VERIFY_FILE_CMD:
			downloadAndVerifyFileCmd()
		case CORRUPT_FILE_CMD:
			corruptFileCmd()
		case EXIT_CMD:
			exitCmd()
		}

		fmt.Println()
	}
}

func createFilesCmd() {
	chDeletingFiles := startLoading("Deleting previous test files")
	deleteTestFiles()
	endLoading(chDeletingFiles)

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

	ch := startLoading(fmt.Sprintf("Creating %d test files", amount))
	start := time.Now()
	fileutil.WriteDummyFiles(TEST_FILE_PATH, amount)
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("%d test files created %s\n\n", amount, elapsed)
}

func createTreeCmd() {
	start := time.Now()
	ch := startLoading("Reading files and hashing data")
	files := fileutil.GetFiles(TEST_FILE_PATH)
	if len(files) < 1 {
		endLoading(ch)
		fmt.Println("Please create some test files first.")
		return
	}

	var fileHashes [][]byte
	for _, file := range files {
		fileHash := sha256.Sum256([]byte(file.Data))
		fileHashes = append(fileHashes, fileHash[:])
	}
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Files hashed %s\n", elapsed)

	ch = startLoading("Building tree")
	start = time.Now()
	root = merkletree.BuildTree(fileHashes)
	rootHash := hex.EncodeToString(root.Hash[:])
	elapsed = time.Since(start)
	endLoading(ch)

	fmt.Printf("Generated Merkle tree %s\n", elapsed)
	fmt.Printf("Root hash: %s\n", rootHash)
}

func uploadFilesCmd() {
	ch := startLoading("Reading test files")
	start := time.Now()
	files := fileutil.GetFiles(TEST_FILE_PATH)
	elapsed := time.Since(start)
	endLoading(ch)

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

	ch = startLoading(fmt.Sprintf("Uploading %d files", len(files)))
	start = time.Now()
	err = api.SendFiles(serverURL, files)
	endLoading(ch)
	if err != nil {
		fmt.Println("Error sending the files to the server:", err)
		return
	}
	elapsed = time.Since(start)
	fmt.Printf("Uploaded %d files! %s\n\n", len(files), elapsed)
}

func downloadAndVerifyFileCmd() {
	if root == nil {
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

	filePath := "./" + DOWNLOAD_FILE_PATH + "/" + fileName
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
	isVerified, proofRoot := merkletree.VerifyMerkleProof(root.Hash, fileHash[:], proof)
	rootHash := hex.EncodeToString(root.Hash)
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

func corruptFileCmd() {
	prompt := promptui.Prompt{
		Label: "Enter file id",
	}
	input, err := prompt.Run()
	if err != nil {
		fmt.Println("Error with prompt:", err)
		return
	}

	start := time.Now()
	file, err := fileutil.GetFile("corrupt.txt")
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

func deleteTestFilesCmd() {
	ch := startLoading("Deleting test files")
	start := time.Now()
	deleteTestFiles()
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Test files deleted! %s\n\n", elapsed)
}

func deleteDownloadsCmd() {
	start := time.Now()
	ch := startLoading("Deleting downloads")
	deleteDownloads()
	elapsed := time.Since(start)
	endLoading(ch)
	fmt.Printf("Downloads deleted! %s\n\n", elapsed)
}

func exitCmd() {
	fmt.Println("Goodbye!")
	os.Exit(0)
}

func deleteTestFiles() {
	fileutil.RemoveDir(TEST_FILE_PATH)
	fileutil.MakeDir(TEST_FILE_PATH)
}

func deleteDownloads() {
	fileutil.RemoveDir(DOWNLOAD_FILE_PATH)
	fileutil.MakeDir(DOWNLOAD_FILE_PATH)
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
