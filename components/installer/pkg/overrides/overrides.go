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
func (o *Overrides) Common() Map {
	res := o.common

	if res == nil {
		return Map{}
	}

	return res
}

//ForComponent returns overrides defined only for specified component
func (o *Overrides) ForComponent(componentName string) Map {

	res := o.components[componentName]

	if res == nil {
		return Map{}
	}

	return res
}

//componentVersions reads overrides for component versions (versions.yaml)
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

//commonOverrides reads overrides common to all components
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

//componentOverrides reads overrides specific for components.
//Returns a map where a key is component name and value is a Map of overrides for the component.
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
