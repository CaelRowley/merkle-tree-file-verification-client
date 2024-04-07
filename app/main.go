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

	const createFilesCmdText = "Create Test Files"
	const createTreeCmdText = "Generate Merkle Tree"
	const uploadFilesCmdText = "Upload Test Files"
	const deleteTestFilesCmdText = "Delete Local Test Files"
	const deleteDownloadCmdText = "Delete Downloads"
	const downloadAndVerifyFileCmdText = "Download and Verify File"
	const corruptFileCmdText = "Corrupt a File on Server"
	const exitCmdText = "Exit"

	items := []string{
		createFilesCmdText,
		createTreeCmdText,
		uploadFilesCmdText,
		downloadAndVerifyFileCmdText,
		corruptFileCmdText,
		deleteTestFilesCmdText,
		deleteDownloadCmdText,
		exitCmdText,
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
		case createFilesCmdText:
			commands.CreateFilesCmd()
		case createTreeCmdText:
			commands.CreateTreeCmd()
		case uploadFilesCmdText:
			commands.UploadFilesCmd(serverURL)
		case deleteTestFilesCmdText:
			commands.DeleteTestFilesCmd()
		case deleteDownloadCmdText:
			commands.DeleteDownloadsCmd()
		case downloadAndVerifyFileCmdText:
			commands.DownloadAndVerifyFileCmd(serverURL)
		case corruptFileCmdText:
			commands.CorruptFileCmd(serverURL)
		case exitCmdText:
			commands.ExitCmd()
		}

		fmt.Println()
		fmt.Println()
	}
}
