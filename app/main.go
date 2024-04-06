package main

import (
	"fmt"
	"log"
	"os"

	"github.com/manifoldco/promptui"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/commands"
)

var serverURL = os.Getenv("SERVER_URL")

func main() {
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	const CREATE_FILES_CMD = "Create Test Files"
	const CREATE_TREE_COMMAND = "Generate Merkle Tree"
	const UPLOAD_FILES_CMD = "Upload Test Files"
	const DELETE_TEST_FILES_CMD = "Delete Local Test Files"
	const DELETE_DOWNLOADS_CMD = "Delete Downloads"
	const DOWNLOAD_AND_VERIFY_FILE_CMD = "Download and Verify File"
	const CORRUPT_FILE_CMD = "Corrupt a File on Server"
	const EXIT_CMD = "Exit"

	items := []string{
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
		Items: items,
		Size:  len(items),
	}

	for {
		_, selected, err := prompt.Run()
		if err != nil {
			log.Fatal(err)
		}

		switch selected {
		case CREATE_FILES_CMD:
			commands.CreateFilesCmd()
		case CREATE_TREE_COMMAND:
			commands.CreateTreeCmd()
		case UPLOAD_FILES_CMD:
			commands.UploadFilesCmd(serverURL)
		case DELETE_TEST_FILES_CMD:
			commands.DeleteTestFilesCmd()
		case DELETE_DOWNLOADS_CMD:
			commands.DeleteDownloadsCmd()
		case DOWNLOAD_AND_VERIFY_FILE_CMD:
			commands.DownloadAndVerifyFileCmd(serverURL)
		case CORRUPT_FILE_CMD:
			commands.CorruptFileCmd(serverURL)
		case EXIT_CMD:
			commands.ExitCmd()
		}

		fmt.Println()
		fmt.Println()
	}
}
