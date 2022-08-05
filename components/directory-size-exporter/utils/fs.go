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

	var newFile *os.File = nil
	newFile, err = os.Create(dirPath + "/" + dirName + "/" + fileName)
	if err != nil {
		return err
	}

	err = os.Truncate(dirPath+"/"+dirName+"/"+fileName, size)
	if err != nil {
		newFile.Close()
		return err
	}
	newFile.Close()

	return err
}
