package pretty

type Kind int

const (
	Unknown Kind = iota
	IDPPreset
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
