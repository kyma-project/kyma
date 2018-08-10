package overrides

import "k8s.io/client-go/kubernetes"

type Overrides struct {
	common     Map
	components map[string]Map
}

//New returns new Overrides instance.
func New(client *kubernetes.Clientset) (*Overrides, error) {

	r := &reader{
		client: client,
	}

	versionsMap, err := componentVersions(r)
	if err != nil {
		return nil, err
	}

	commonsMap, err := commonOverrides(r)
	if err != nil {
		return nil, err
	}

	commonOverridesMap := Map{}
	MergeMaps(commonOverridesMap, versionsMap)
	MergeMaps(commonOverridesMap, commonsMap)

	componentsMap, err := componentOverrides(r)
	if err != nil {
		return nil, err
	}

	res := Overrides{
		common:     commonOverridesMap,
		components: componentsMap,
	}

	return &res, nil
}

//Common returns overrides common for all components
func (o *Overrides) Common(componentName string) (Map, error) {
	res := o.components[componentName]

	if res == nil {
		return Map{}, nil
	}

	return res, nil
}

//OverridesFor returns overrides defined only for specified component
func (o *Overrides) OnlyFor(componentName string) (Map, error) {

	specific := o.components[componentName]
	if specific == nil {
		return Map{}, nil
	}

	return specific, nil
}

func componentVersions(r *reader) (Map, error) {

	versionsFileData, err := loadComponentsVersions()
	if err != nil {
		return nil, err
	}

	if versionsFileData == nil {
		return Map{}, nil
	}

	return ToMap(versionsFileData.String())
}

func commonOverrides(r *reader) (Map, error) {
	common, err := r.getCommonConfig()
	if err != nil {
		return nil, err
	}

	if common == nil {
		return Map{}, nil
	}

	return UnflattenToMap(common), nil
}

func componentOverrides(r *reader) (map[string]Map, error) {

	components, err := r.getComponents()
	if err != nil {
		return nil, err
	}

	res := map[string]Map{}
	if components == nil {
		return res, nil
	}

	for _, c := range components {
		unflattened := UnflattenToMap(c.overrides)
		res[c.name] = unflattened
	}

	return res, nil
}
