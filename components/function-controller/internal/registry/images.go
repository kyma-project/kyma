package registry

type ImageList map[string]map[string]struct{}

func NewTaggedImageList() ImageList {
	return ImageList{}
}

func (f ImageList) AddImageWithTag(image, tag string) {
	if _, ok := f[image]; !ok {
		f[image] = map[string]struct{}{}
	}
	f[image][tag] = struct{}{}
}

func (f ImageList) HasImageWithTag(image, tag string) bool {
	tags, ok := f[image]
	if !ok {
		return false
	}
	_, hasTag := tags[tag]
	return hasTag
}

func (f ImageList) ListImages() []string {
	images := []string{}
	for k := range f {
		images = append(images, k)
	}
	return images
}
