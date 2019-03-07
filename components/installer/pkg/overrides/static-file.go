package overrides

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

// StaticFile interface defines contract for overrides file representation
type StaticFile interface {
	HasOverrides() bool
	GetOverrides() (Map, error)
}

// LocalStaticFile struct defines static file overrides for local
type LocalStaticFile struct{}

// NewLocalStaticFile function creates instance of LocalStaticFile struct for cluster overrides
func NewLocalStaticFile() *LocalStaticFile {
	return &LocalStaticFile{}
}

// HasOverrides .
func (localStaticFile *LocalStaticFile) HasOverrides() bool {
	return false
}

// GetOverrides .
func (localStaticFile *LocalStaticFile) GetOverrides() (Map, error) {
	return Map{}, nil
}

// ClusterStaticFile struct defines static file overrides for cluster
type ClusterStaticFile struct {
	DirectoryPath *string
}

// NewClusterStaticFile function creates instance of ClusterStaticFile struct for cluster overrides
func NewClusterStaticFile(dirPath string) *ClusterStaticFile {
	return &ClusterStaticFile{
		DirectoryPath: &dirPath,
	}
}

// HasOverrides function returns boolean whether additional overrides are defined
func (clusterStaticFile *ClusterStaticFile) HasOverrides() bool {
	if clusterStaticFile.DirectoryPath == nil {
		return false
	}

	if _, err := os.Stat(clusterStaticFile.getFilePath()); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetOverrides function reads cluster overrides file and returns its content
func (clusterStaticFile *ClusterStaticFile) GetOverrides() (Map, error) {
	fileBytes, err := ioutil.ReadFile(clusterStaticFile.getFilePath())

	if err != nil {
		log.Printf(
			"An error occured while reading file with additional overrides from path %s",
			clusterStaticFile.getFilePath())

		return nil, err
	}

	return ToMap(string(fileBytes))
}

func (clusterStaticFile *ClusterStaticFile) getFilePath() string {
	const clusterStaticFileName = "cluster.yaml"

	return path.Join(*clusterStaticFile.DirectoryPath, clusterStaticFileName)
}
