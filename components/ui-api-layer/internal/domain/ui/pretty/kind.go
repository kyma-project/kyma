package pretty

type Kind int

const (
	IDPPreset Kind = iota
	IDPPresets
)

func (k Kind) String() string {
	switch k {
	case IDPPreset:
		return "IDP Preset"
	case IDPPresets:
		return "IDP Presets"
	default:
		return ""
	}
}
