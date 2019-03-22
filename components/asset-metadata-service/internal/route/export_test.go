package route

import "github.com/kyma-project/kyma/components/asset-metadata-service/pkg/extractor"

func (h *ExtractHandler) SetMetadataExtractor(metadataExtractor extractor.Extractor) {
	h.metadataExtractor = metadataExtractor
}
