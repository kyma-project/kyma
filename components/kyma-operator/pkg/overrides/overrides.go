package overrides

import (
	"k8s.io/client-go/kubernetes"
)

//OverrideData exposes methods to fetch release-specific overrides
type OverrideData interface {
	ForRelease(releaseName string) (string, error)
}

//Provider implements storage for all overrides
type Provider struct {
	common       Map
	components   map[string]Map
	configReader reader
}

//ForRelease returns overrides for release
func (o *Provider) ForRelease(releaseName string) (string, error) {

	o.refreshStore()
	allOverrides := Map{}

	MergeMaps(allOverrides, o.common)
	MergeMaps(allOverrides, o.components[releaseName])

	return ToYaml(allOverrides)
}

//New returns new Data instance.
func New(client kubernetes.Interface) OverrideData {

	configReader := &reader{
		client: client,
	}

	res := Provider{
		configReader: *configReader,
	}

	return &res
}

func (o *Provider) refreshStore() error {

	versionsMap, err := versionOverrides()
	if err != nil {
		return err
	}

	commonOverridesData, err := o.configReader.readCommonOverrides()
	if err != nil {
		return err
	}

	commonOverrides := UnflattenToMap(joinOverridesMap(commonOverridesData...))
	commonOverridesMap := Map{}
	MergeMaps(commonOverridesMap, versionsMap)
	MergeMaps(commonOverridesMap, commonOverrides)

	componentsOverridesData, err := o.configReader.readComponentOverrides()
	if err != nil {
		return err
	}

	componentsMap := unflattenComponentOverrides(joinComponentOverrides(componentsOverridesData...))

	o.common = commonOverridesMap
	o.components = componentsMap

	return nil
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
