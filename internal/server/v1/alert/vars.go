package alert

import "golang.org/x/exp/maps"

var SuppliedVariables = []string{"name", "team", "entity"}

func RemoveSuppliedVariablesFromTemplates(templates []Template, varKeys []string) []Template {
	var result []Template
	for _, t := range templates {
		var finalVars []Variable
		for _, variable := range t.Variables {
			if !contains(varKeys, variable.Name) {
				finalVars = append(finalVars, variable)
			}
		}
		t.Variables = finalVars
		result = append(result, t)
	}
	return result
}

func RemoveSuppliedVariablesFromRules(rules []Rule, varKeys []string) []Rule {
	var result []Rule
	for _, r := range rules {
		var finalVars []Variable
		for _, variable := range r.Variables {
			if !contains(varKeys, variable.Name) {
				finalVars = append(finalVars, variable)
			}
		}
		r.Variables = finalVars
		result = append(result, r)
	}
	return result
}

func AddSuppliedVariablesFromRules(rules []Rule, vars map[string]string) []Rule {
	rules = RemoveSuppliedVariablesFromRules(rules, maps.Keys(vars))
	var suppliedVars []Variable
	for k, v := range vars {
		suppliedVars = append(suppliedVars, Variable{
			Name:  k,
			Value: v,
		})
	}

	var result []Rule
	for _, rule := range rules {
		rule.Variables = append(rule.Variables, suppliedVars...)
		result = append(result, rule)
	}
	return result
}

func contains(arr []string, item string) bool {
	for _, arrItem := range arr {
		if arrItem == item {
			return true
		}
	}
	return false
}
