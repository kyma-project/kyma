package automock

func NewIDPPresetSvc() *idpPresetSvc {
	return new(idpPresetSvc)
}

func NewGQLIDPPresetConverter() *gqlIDPPresetConverter {
	return new(gqlIDPPresetConverter)
}
