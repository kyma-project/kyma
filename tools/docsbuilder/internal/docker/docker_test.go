package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageName(t *testing.T) {
	testCases := []struct {
		docName  string
		cfg      *Config
		expected string
	}{
		{
			docName: "test",
			cfg: &Config{
				ImageTag: "latest",
			},
			expected: "test:latest",
		},
		{
			docName: "test2",
			cfg: &Config{
				ImageTag:    "foo",
				ImageSuffix: "-docs",
			},
			expected: "test2-docs:foo",
		},
		{
			docName: "test3",
			cfg: &Config{
				ImageTag:    "bar",
				ImageSuffix: "-suffix",
				ImagePrefix: "repository.example.com/",
			},
			expected: "repository.example.com/test3-suffix:bar",
		},
	}

	for _, tC := range testCases {
		result := ImageName(tC.docName, *tC.cfg)
		assert.Equal(t, tC.expected, result)
	}
}

func TestBuildCommand(t *testing.T) {
	expected := "cat /path/to/Dockerfile | docker build -f - . -t example.com/test:latest --example=test"

	result := buildCommand(&ImageBuildConfig{
		Name:                "example.com/test:latest",
		DockerfilePath:      "/path/to/Dockerfile",
		AdditionalBuildArgs: "--example=test",
	})

	assert.Equal(t, expected, result)
}

func TestPushCommand(t *testing.T) {
	expected := "docker push example.com/test:latest"

	result := pushCommand("example.com/test:latest")

	assert.Equal(t, expected, result)
}
