package config

type Parameter struct {
	Key   string
	Value string
}

type ParameterList []Parameter

func (p *ParameterList) ContainsKey(key string) bool {
	for _, param := range *p {
		if param.Key == key {
			return true
		}
	}
	return false
}

func (p *ParameterList) Add(parameter Parameter) {
	result := *p
	result = append(result, parameter)
	*p = result
}

func (p *ParameterList) GetByKey(key string) *Parameter {
	for _, param := range *p {
		if param.Key == key {
			return &param
		}
	}
	return nil
}
