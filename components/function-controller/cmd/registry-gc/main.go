package main

import (
	"context"

	"github.com/kyma-project/kyma/components/function-controller/internal/registry"
)

func main() {
	repoCli, err := registry.NewRepositoryClient(context.Background(),
		registry.RepositoryClientOptions{
			Image: "scratch",
			URL:   "http://localhost:5000",
		})
	if err != nil {
		panic(err)
	}

	tagList, err := repoCli.ListTags()
	if err != nil {
		panic(err)
	}
	for _, tagStr := range tagList {
		tag, err := repoCli.GetImageTag(tagStr)
		if err != nil {
			panic(err)
		}
		err = repoCli.DeleteImageTag(tag.Digest)
		if err != nil {
			panic(err)
		}
	}

}
