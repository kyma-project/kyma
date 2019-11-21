package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Response struct {
	UploadedFiles []Result `json:"uploadedFiles,omitempty"`
	Errors        []Error         `json:"errors,omitempty"`
}

type Result struct {
	FileName   string `json:"fileName"`
	RemotePath string `json:"remotePath"`
	Bucket     string `json:"bucket"`
	Size       int64  `json:"size"`
}

type Error struct {
	Message  string `json:"message"`
	FileName string `json:"omitempty,fileName"`
}

type UploadInput struct {
	PrivateFiles []*os.File
	PublicFiles []*os.File
	Directory string
}

func Do(directory string, input UploadInput, url string) (*Response, error) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)

	for _, file := range input.PrivateFiles {
		if file == nil {
			continue
		}

		w, err := formWriter.CreateFormFile("private", file.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "while creating form field %s to buffer", file.Name())
		}

		_, err = io.Copy(w, file)
		if err != nil {
			return nil, errors.Wrapf(err, "while copying file %s to buffer", file.Name())
		}
	}

	for _, file := range input.PublicFiles {
		if file == nil {
			continue
		}

		w, err := formWriter.CreateFormFile("public", filepath.Base(file.Name()))
		if err != nil {
			return nil, errors.Wrapf(err, "while creating form field %s to buffer", file.Name())
		}

		_, err = io.Copy(w, file)
		if err != nil {
			return nil, errors.Wrapf(err, "while copying file %s to buffer", file.Name())
		}
	}

	if directory != "" {
		err := formWriter.WriteField("directory", directory)
		if err != nil {
			return nil, errors.Wrap(err, "while creating field in form")
		}
	}
	err := formWriter.Close()
	if err != nil {
		return nil, errors.Wrap(err, "while closing form")
	}

	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return nil, errors.Wrap(err, "while creating request")
	}

	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	result, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "while doing upload request")
	}

	defer func() {
		err := result.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Upload status code: %d. Expected %d", result.StatusCode, http.StatusOK)
	}

	var response Response
	err = json.NewDecoder(result.Body).Decode(&response)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding response")
	}

	return &response, nil
}