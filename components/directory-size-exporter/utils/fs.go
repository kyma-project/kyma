package utils

import "os"

func PrepareMockDirectories(testDir string) (string, error) {
	dirPath := testDir + "/test-data"
	err := os.Mkdir(dirPath, 0700)
	if err != nil {
		return "", err
	}

	emitters := []string{"emitter1", "emitter2", "emitter3"}
	for i, emitterName := range emitters {
		err = PrepareMockDirectory(dirPath, emitterName, int64(i*100))
		if err != nil {
			return "", err
		}
	}

	return dirPath, err
}

func PrepareMockDirectory(dirPath string, dirName string, size int64) error {
	const fileName string = "test.txt"

	err := os.Mkdir(dirPath+"/"+dirName, 0700)
	if err != nil {
		return err
	}

	_, err = WriteMockFileToDirectory(dirPath+"/"+dirName, fileName, size)

	return err
}

func WriteMockFileToDirectory(dirPath string, filename string, size int64) (*os.File, error) {
	var newFile *os.File
	newFile, err := os.Create(dirPath + "/" + filename)
	if err != nil {
		return nil, err
	}

	err = os.Truncate(dirPath+"/"+filename, size)
	if err != nil {
		newFile.Close()
		return nil, err
	}

	newFile.Close()

	return newFile, err
}
