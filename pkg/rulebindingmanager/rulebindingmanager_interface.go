package rulebindingmanager

import (
	"github.com/Aryaman6492/node-agent/pkg/ruleengine"
)

type RuleBindingCache interface {
	ListRulesForPod(namespace, name string) []ruleengine.RuleEvaluator
	AddNotifier(*chan RuleBindingNotify)
	GetRuleCreator() ruleengine.RuleCreator
}
