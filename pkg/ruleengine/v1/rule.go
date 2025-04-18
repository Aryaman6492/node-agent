package ruleengine

import (
	"github.com/Aryaman6492/node-agent/pkg/ruleengine"
	"github.com/Aryaman6492/node-agent/pkg/utils"

	"github.com/goradd/maps"
)

const (
	RulePriorityNone        = 0
	RulePriorityLow         = 1
	RulePriorityMed         = 5
	RulePriorityHigh        = 8
	RulePriorityCritical    = 10
	RulePrioritySystemIssue = 1000
)

var _ ruleengine.RuleSpec = (*RuleRequirements)(nil)

type RuleRequirements struct {
	// Needed events for the rule.
	EventTypes []utils.EventType
}

// Event types required for the rule
func (r *RuleRequirements) RequiredEventTypes() []utils.EventType {
	return r.EventTypes
}

type BaseRule struct {
	// Mutex for protecting rule parameters.
	parameters maps.SafeMap[string, interface{}]
}

func (br *BaseRule) SetParameters(parameters map[string]interface{}) {
	for k, v := range parameters {
		br.parameters.Set(k, v)
	}
}

func (br *BaseRule) GetParameters() map[string]interface{} {

	// Create a copy to avoid returning a reference to the internal map
	parametersCopy := make(map[string]interface{}, br.parameters.Len())

	br.parameters.Range(
		func(key string, value interface{}) bool {
			parametersCopy[key] = value
			return true
		},
	)
	return parametersCopy
}
