package compass

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type ScenariosSchema struct {
	Type        string         `json:"type"`
	MinItems    int            `json:"minItems"`
	UniqueItems bool           `json:"uniqueItems"`
	Items       ScenariosItems `json:"items"`
}

type ScenariosItems struct {
	Type string   `json:"type"`
	Enum []string `json:"enum"`
}

func ToScenarioSchema(scenarioLabelDefinition ScenarioLabelDefinition) (ScenariosSchema, error) {
	var scenarioSchema ScenariosSchema

	if scenarioLabelDefinition.Schema == nil {
		return ScenariosSchema{}, nil
	}

	err := json.Unmarshal([]byte(*scenarioLabelDefinition.Schema), &scenarioSchema)
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to unmarshall scenario schema")
	}
	return scenarioSchema, nil
}

func (ss *ScenariosSchema) ToLabelDefinitionInput(key string) (graphql.LabelDefinitionInput, error) {
	var inputSchema interface{} = ss
	schema, err := graphql.MarshalSchema(&inputSchema)
	if err != nil {
		return graphql.LabelDefinitionInput{}, err
	}

	return graphql.LabelDefinitionInput{
		Key:    key,
		Schema: schema,
	}, nil
}

func (ss *ScenariosSchema) AddScenario(value string) {
	ss.Items.Enum = append(ss.Items.Enum, value)
}

func (ss *ScenariosSchema) RemoveScenario(value string) {
	if ss.Items.Enum == nil {
		return
	}

	scenarios := make([]string, 0)
	for _, s := range ss.Items.Enum {
		if s != value {
			scenarios = append(scenarios, s)
		}
	}

	ss.Items.Enum = scenarios
}
