package overrides

import "k8s.io/client-go/kubernetes"

type Data struct {
	common     Map
	components map[string]Map
}

//New returns new Data instance.
func New(client *kubernetes.Clientset) (*Data, error) {

	r := &reader{
		client: client,
	}

	versionsMap, err := versionOverrides()
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

	res := Data{
		common:     commonOverridesMap,
		components: componentsMap,
	}

	return &res, nil
}

//Common returns overrides common for all components
func (o *Data) Common() Map {
	res := o.common

	if res == nil {
		return Map{}
	}

	return res
}

//ForComponent returns overrides defined only for specified component
func (o *Data) ForComponent(componentName string) Map {

	res := o.components[componentName]

	if res == nil {
		return Map{}
	}

	return res
}

//versionOverrides reads overrides for component versions (versions.yaml)
func versionOverrides() (Map, error) {

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
