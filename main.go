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
	deleteTestFiles()
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

	fmt.Printf("Creating %d test files...\n", amount)
	start := time.Now()
	fileutil.WriteDummyFiles(TEST_FILE_PATH, amount)
	elapsed := time.Since(start)
	fmt.Printf("%d test files created %s\n\n", amount, elapsed)
}

func createTreeCmd() {
	files := fileutil.GetFiles(TEST_FILE_PATH)
	if len(files) < 1 {
		fmt.Println("Please create some test files first.")
		return
	}

	var fileHashes [][]byte
	for _, file := range files {
		fileHash := sha256.Sum256([]byte(file.Data))
		fileHashes = append(fileHashes, fileHash[:])
	}

	fmt.Printf("Building Merkle tree...\n")
	start := time.Now()
	root = merkletree.BuildTree(fileHashes)
	rootHash := hex.EncodeToString(root.Hash[:])
	elapsed := time.Since(start)
	fmt.Printf("Generated Merkle tree with root hash: %s %s\n", rootHash, elapsed)
}

func uploadFilesCmd() {
	fmt.Printf("Reading test files...\n")
	start := time.Now()
	files := fileutil.GetFiles(TEST_FILE_PATH)
	elapsed := time.Since(start)
	fmt.Printf("%d test files read %s\n\n", len(files), elapsed)

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

	fmt.Printf("Uploading %d files...\n", len(files))
	start = time.Now()
	err = api.SendFiles(serverURL, files)
	if err != nil {
		fmt.Println("Error sending the files to the server:", err)
		return
	}
	elapsed = time.Since(start)
	fmt.Printf("Uploaded %d files! %s\n\n", len(files), elapsed)
}

func deleteTestFilesCmd() {
	fmt.Println("Deleting all test files...")
	start := time.Now()
	deleteTestFiles()
	elapsed := time.Since(start)
	fmt.Printf("Test files deleted %s\n\n", elapsed)
}

func deleteDownloadsCmd() {
	fmt.Println("Deleting all  downloads...")
	start := time.Now()
	deleteDownloads()
	elapsed := time.Since(start)
	fmt.Printf("Downloads deleted %s\n\n", elapsed)
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

	file, err := api.GetFile(serverURL, input, DOWNLOAD_FILE_PATH)
	if err != nil {
		fmt.Println("Error getting file with id:", input, ":", err)
		return
	}
	proof, err := api.GetProof(serverURL, input)
	if err != nil {
		fmt.Println("Error getting proof:", err)
		return
	}

	fileHash := sha256.Sum256(file)
	fmt.Printf("Downloaded file: %s\n", input)

	if merkletree.VerifyMerkleProof(root.Hash, fileHash[:], proof) {
		fmt.Print("Merkle proof verification successful!\nFile integrity is confirmed.\n\n")
	} else {
		fmt.Print("Merkle proof verification failed!\nFile integrity cannot be confirmed.\n\n")
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
	fmt.Printf("Corrupted file: %s\n\n", input)
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
