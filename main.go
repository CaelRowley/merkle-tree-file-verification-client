package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/manifoldco/promptui"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/api"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/merkletree"
)

const FILE_PATH = "testfiles"
const DOWNLOAD_PATH = "downloads"

var (
	serverURL = os.Getenv("SERVER_URL")
)

var root *merkletree.Node

func main() {
	cleanFiles()
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	const UPLOAD_FILES_CMD = "Upload files"
	const DOWNLOAD_AND_VERIFY_FILE_CMD = "Download and verify file"
	const CORRUPT_FILE_CMD = "Corrupt a file"
	const EXIT_CMD = "Exit"

	commands := []string{
		UPLOAD_FILES_CMD,
		DOWNLOAD_AND_VERIFY_FILE_CMD,
		CORRUPT_FILE_CMD,
		EXIT_CMD,
	}

	prompt := promptui.Select{
		Label: "Select a command",
		Items: commands,
	}

	for {
		_, selected, err := prompt.Run()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println()

		switch selected {
		case UPLOAD_FILES_CMD:
			uploadFilesCmd()
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

func uploadFilesCmd() {
	prompt := promptui.Prompt{
		Label: "Amount to upload",
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

	fileutil.WriteDummyFiles(FILE_PATH, amount)
	files := fileutil.GetFiles(FILE_PATH)
	var fileHashes [][]byte
	for _, file := range files {
		fileHash := sha256.Sum256([]byte(file.Data))
		fileHashes = append(fileHashes, fileHash[:])
	}
	fmt.Printf("Created %d files!\n", amount)

	root = merkletree.BuildTree(fileHashes)
	rootHash := hex.EncodeToString(root.Hash[:])
	fmt.Printf("Generated merkle tree with root hash: %s\n", rootHash)

	err = api.DeleteAllFiles(serverURL)
	if err != nil {
		fmt.Println("Error deleting files in the DB:", err)
		return
	}

	fmt.Println("Deleted all files in the DB!")

	err = api.SendFiles(serverURL, files)
	if err != nil {
		fmt.Println("Error sending the files to the server:", err)
		return
	}
	fmt.Printf("Uploaded %d files!\n", amount)

	fileutil.RemoveDir(FILE_PATH)
	fileutil.MakeDir(FILE_PATH)

	fmt.Println("Deleted all local files!")
}

func downloadAndVerifyFileCmd() {
	if root == nil {
		fmt.Println("You need to upload files first")
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

	file, err := api.GetFile(serverURL, input, DOWNLOAD_PATH)
	if err == nil {
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
		fmt.Println("Merkle proof verification successful!\nFile integrity is confirmed.")
	} else {
		fmt.Println("Merkle proof verification failed!\nFile integrity cannot be confirmed.")
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
	fmt.Printf("Corrupted file: %s\n", input)
}

func exitCmd() {
	cleanFiles()
	fmt.Println("Goodbye!")
	os.Exit(0)
}

func cleanFiles() {
	fileutil.RemoveDir(FILE_PATH)
	fileutil.MakeDir(FILE_PATH)
	fileutil.RemoveDir(DOWNLOAD_PATH)
	fileutil.MakeDir(DOWNLOAD_PATH)
}
