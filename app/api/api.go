package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/google/uuid"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-client/app/utils/merkletree"
)

func UploadFiles(url string, files []fileutil.File) error {
	batchSize := 10000

	batchId, err := uuid.NewV7()
	if err != nil {
		return err
	}

	requestUrl := fmt.Sprintf("%s/files/upload-batch/%s", url, batchId)

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		jsonData, err := json.Marshal(batch)
		if err != nil {
			return err
		}
		res, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("server responded with non-OK status: %d", res.StatusCode)
		}
	}

	return nil
}

func GetProof(url string, id string) (merkletree.MerkleProof, error) {
	requestUrl := fmt.Sprintf("%s/files/get-proof/%s", url, id)
	res, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
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
	res, err := http.Get(requestUrl)
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return "", nil, errors.New("No file found for id: " + id)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}

	if len(res.Header["Content-Disposition"]) == 0 {
		return "", nil, errors.New("filename missing from header")
	}

	_, params, err := mime.ParseMediaType(res.Header["Content-Disposition"][0])
	if err != nil {
		return "", nil, err
	}
	fileName := params["filename"]

	return fileName, body, nil
}

func DeleteAllFiles(url string) error {
	requestUrl := url + "/files/delete-all"
	res, err := http.Post(requestUrl, "application/json", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func CorruptFile(url string, id string, file []byte) error {
	requestUrl := fmt.Sprintf("%s/files/corrupt-file/%s", url, id)
	jsonData, err := json.Marshal(file)
	if err != nil {
		return err
	}

	res, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server responded with non-OK status: %d", res.StatusCode)
	}

	return nil
}
