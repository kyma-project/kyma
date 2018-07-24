package main

import (
	"io"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type index struct {
	APIVersion string                               `json:"apiVersion"`
	Entries    map[internal.BundleName][]indexEntry `json:"entries"`
}

type indexEntry struct {
	Name        internal.BundleName `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"version"`
}

func render(in []*internal.Bundle, w io.Writer) error {
	dto := &index{
		APIVersion: "v1",
		Entries:    make(map[internal.BundleName][]indexEntry),
	}

	for _, b := range in {
		e := indexEntry{
			Name:        b.Name,
			Description: b.Description,
			Version:     b.Version.String(),
		}
		dto.Entries[b.Name] = append(dto.Entries[b.Name], e)
	}

	entEnc, err := yaml.Marshal(dto)
	if err != nil {
		return errors.Wrap(err, "while encoding to YAML")
	}

	if _, err := w.Write(entEnc); err != nil {
		return errors.Wrap(err, "while writing encoded index")
	}

	return nil
}
