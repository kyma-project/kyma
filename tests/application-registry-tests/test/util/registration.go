package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type serviceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var httpClient http.Client = http.Client{}

func DeleteAll(registryAPIURL string) error {

	if registryAPIURL[len(registryAPIURL)-1] == "/"[0] {
		registryAPIURL = registryAPIURL[:len(registryAPIURL)-1]
	}
	registeredServices, err := readAll(registryAPIURL)

	if err != nil {
		return err
	}

	for i := range registeredServices {
		request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", registryAPIURL, registeredServices[i]), nil)
		if err != nil {
			return err
		}

		_, err = httpClient.Do(request)
		if err != nil {
			return err
		}
	}

	return nil
}

func readAll(registryAPIURL string) (registerdAPIs []string, err error) {

	response, err := httpClient.Get(registryAPIURL)
	if err != nil {
		return
	}
	defer response.Body.Close()

	var services []serviceResponse

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(responseBytes, &services)
	if err != nil {
		return
	}

	registerdAPIs = make([]string, len(services))

	for i := range services {
		registerdAPIs[i] = services[i].ID
	}

	return
}

func RegisterSpec(registryAPIURL string, spec []byte) error {

	if registryAPIURL[len(registryAPIURL)-1] == "/"[0] {
		registryAPIURL = registryAPIURL[:len(registryAPIURL)-1]
	}

	request, err := http.NewRequest(http.MethodPost, registryAPIURL, bytes.NewReader(spec))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	_, err = httpClient.Do(request)

	return err
}

func RegisterSpecForFolder(registryAPIURL string, folderPath string) error {

	err := DeleteAll(registryAPIURL)

	if err != nil {
		return fmt.Errorf("error unregistering services: %s", err.Error())
	}

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error opening sample file %q: %s", path, err.Error())
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".json" {
			filePtr, err := os.Open(path)
			defer filePtr.Close()
			if err != nil {
				return fmt.Errorf("error opening sample file %q: %s", path, err.Error())
			}

			specBytes, err := ioutil.ReadAll(filePtr)
			if err != nil {
				return fmt.Errorf("error reading sample file %q: %s", path, err.Error())
			}

			err = RegisterSpec(registryAPIURL, specBytes)
			if err != nil {
				return fmt.Errorf("error registering sample file %q: %s", path, err.Error())
			}
		}

		return nil
	})

	return err
}
