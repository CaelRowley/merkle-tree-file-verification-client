package main

import (
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
)

func main() {
	const UPLOAD_FILES_CMD = "Upload files"
	const DOWNLOAD_AND_VERIFY_FILE_CMD = "Download and verify file"

	for {
		commands := []string{
			UPLOAD_FILES_CMD,
			DOWNLOAD_AND_VERIFY_FILE_CMD,
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
			uploadFiles()
		case DOWNLOAD_AND_VERIFY_FILE_CMD:
			downloadAndVerifyFile()
		}
	}
}

func uploadFiles() {
	var text string
	prompt := promptui.Prompt{
		Label: "Amount to upload",
	}
	text, err := prompt.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Uploaded %s files!\n\n", text)
}

func downloadAndVerifyFile() {
	var text string
	prompt := promptui.Prompt{
		Label: "Enter file id",
	}
	text, err := prompt.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Downloaded and verified file: %s\n\n", text)
}
