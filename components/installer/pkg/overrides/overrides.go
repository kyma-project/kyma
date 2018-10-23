package overrides

import (
	"k8s.io/client-go/kubernetes"
)

type OverrideData interface {
	Common() Map
	ForComponent(componentName string) Map
	ForRelease(releaseName string) (string, error)
}

type Provider struct {
	common     Map
	components map[string]Map
}

//Common returns overrides common for all components
func (o *Provider) Common() Map {
	res := o.common

	if res == nil {
		return Map{}
	}

	return res
}

//ForComponent returns overrides defined only for specified component
func (o *Provider) ForComponent(componentName string) Map {

	res := o.components[componentName]

	if res == nil {
		return Map{}
	}

	return res
}

//ForRelease returns overrides for release
func (o *Provider) ForRelease(releaseName string) (string, error) {
	allOverrides := Map{}

	MergeMaps(allOverrides, o.Common())
	MergeMaps(allOverrides, o.ForComponent(releaseName))

	return ToYaml(allOverrides)
}

//New returns new Data instance.
func New(client *kubernetes.Clientset) (OverrideData, error) {

	r := &reader{
		client: client,
	}

	versionsMap, err := versionOverrides()
	if err != nil {
		return nil, err
	}

	commonOverridesData, err := r.readCommonOverrides()
	if err != nil {
		return nil, err
	}

	commonOverrides := UnflattenToMap(joinOverridesMap(commonOverridesData...))
	commonOverridesMap := Map{}
	MergeMaps(commonOverridesMap, versionsMap)
	MergeMaps(commonOverridesMap, commonOverrides)

	componentsOverridesData, err := r.readComponentOverrides()
	if err != nil {
		return nil, err
	}

	componentsMap := unflattenComponentOverrides(joinComponentOverrides(componentsOverridesData...))

	res := Provider{
		common:     commonOverridesMap,
		components: componentsMap,
	}

	return &res, nil
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

//joinOverridesMap joins overrides from multiple input maps.
//Maps are joined in order given by input slice, using "last wins" strategy.
func joinOverridesMap(inputMaps ...inputMap) inputMap {

	combined := make(inputMap)

	if inputMaps == nil {
		return combined
	}

	for _, im := range inputMaps {
		for key, val := range im {
			//No error if key's already there, "last wins"
			combined[key] = val
		}
	}

	return combined
}

//joinComponentOverrides joins overrides for components. Useful when multiple input data objects for a single component exist.
//Accepts component slice where overrides for the same component may occur several times.
//Returns a map of input data for components overrides: key is component name, value contains joined overrides for the component.
func joinComponentOverrides(components ...component) map[string]inputMap {

	res := make(map[string]inputMap)

	if components == nil {
		return res
	}

	for _, c := range components {
		if res[c.name] != nil {
			m1 := res[c.name]
			m2 := c.overrides
			res[c.name] = joinOverridesMap(m1, m2)
		} else {
			res[c.name] = c.overrides
		}
	}

	return res
}

//Helper func that unflattens values from input map.
func unflattenComponentOverrides(inputDataMap map[string]inputMap) map[string]Map {

	res := make(map[string]Map)

	if inputDataMap == nil {
		return res
	}

	for key, val := range inputDataMap {
		res[key] = UnflattenToMap(val)
	}

	return res
}
