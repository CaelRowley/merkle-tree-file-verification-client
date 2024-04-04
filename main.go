package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"

	"github.com/manifoldco/promptui"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/merkletree"
)

const FILE_PATH = "testfiles"
const DOWNLOAD_PATH = "downloads"

var root *merkletree.Node

func main() {
	const UPLOAD_FILES_CMD = "Upload files"
	const DOWNLOAD_AND_VERIFY_FILE_CMD = "Download and verify file"
	const EXIT_CMD = "Exit"

	fileutil.RemoveDir(FILE_PATH)
	fileutil.MakeDir(FILE_PATH)
	fileutil.RemoveDir(DOWNLOAD_PATH)
	fileutil.MakeDir(DOWNLOAD_PATH)

	for {
		commands := []string{
			UPLOAD_FILES_CMD,
			DOWNLOAD_AND_VERIFY_FILE_CMD,
			EXIT_CMD,
		}

		prompt := promptui.Select{
			Label: "Select a command",
			Items: commands,
		}

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
		fmt.Println(err)
		return
	}

	amount, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println(err)
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

	err = deleteAllFiles()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Deleted all files in the DB!")

	err = sendFiles(files)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return
	}

	file, err := getFile(input, DOWNLOAD_PATH)
	if err != nil {
		fmt.Println(err)
		return
	}
	proof, err := getProof(input)
	if err != nil {
		fmt.Println(err)
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

func exitCmd() {
	fmt.Println("Goodbye!")
	os.Exit(0)
}

func sendFiles(files []fileutil.File) error {
	requestUrl := "http://localhost:8080/files/upload"
	jsonData, err := json.Marshal(files)
	if err != nil {
		return err
	}

	resp, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server responded with non-OK status: %d", resp.StatusCode)
	}

	return nil
}

func getProof(id string) (merkletree.MerkleProof, error) {
	requestUrl := fmt.Sprintf("http://localhost:8080/files/get-proof/%s", id)
	response, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var proof merkletree.MerkleProof
	err = json.Unmarshal(body, &proof)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func getFile(id string, path string) ([]byte, error) {
	requestUrl := fmt.Sprintf("http://localhost:8080/files/download/%s", id)

	response, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, errors.New("No file found for id: " + id)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	_, params, err := mime.ParseMediaType(response.Header["Content-Disposition"][0])
	if err != nil {
		return nil, err
	}
	filename := params["filename"]

	err = os.WriteFile(path+"/"+filename, body, 0644)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func deleteAllFiles() error {
	requestUrl := "http://localhost:8080/files/delete-all"
	response, err := http.Post(requestUrl, "application/json", nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}
