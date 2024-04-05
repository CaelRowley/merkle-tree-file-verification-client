package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/utils/merkletree"
)

func SendFiles(url string, files []fileutil.File) error {
	requestUrl := url + "/files/upload"
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

func GetProof(url string, id string) (merkletree.MerkleProof, error) {
	requestUrl := fmt.Sprintf("%s/files/get-proof/%s", url, id)
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

func GetFile(url string, id string) (string, []byte, error) {
	requestUrl := fmt.Sprintf("%s/files/download/%s", url, id)

	response, err := http.Get(requestUrl)
	if err != nil {
		return "", nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return "", nil, errors.New("No file found for id: " + id)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", nil, err
	}

	_, params, err := mime.ParseMediaType(response.Header["Content-Disposition"][0])
	if err != nil {
		return "", nil, err
	}
	fileName := params["filename"]

	return fileName, body, nil
}

func DeleteAllFiles(url string) error {
	requestUrl := url + "/files/delete-all"
	response, err := http.Post(requestUrl, "application/json", nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func CorruptFile(url string, id string, file []byte) error {
	requestUrl := fmt.Sprintf("%s/files/corrupt-file/%s", url, id)
	jsonData, err := json.Marshal(file)
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
