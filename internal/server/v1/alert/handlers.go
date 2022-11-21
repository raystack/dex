package firehose

import (
	"net/http"

	"github.com/odpf/dex/internal/server/utils"
)

const (
	alertPolicyNotFound      = "no Alert Policy found for given resource"
	alertProviderName        = "cortex"
	projectSlugSirenLabelKey = "projects"
)

type listResponse[T any] struct {
	Items []T `json:"items"`
}

func HandleListAlertTemplates(svc *Service, tag string, suppliedAlertVariableNames []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		templates, err := svc.ListAlertTemplates(ctx, tag)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		templates = removeSuppliedVariablesFromTemplates(templates, suppliedAlertVariableNames)

		resp := listResponse[Template]{Items: templates}
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}

func removeSuppliedVariablesFromTemplates(templates []Template, varKeys []string) []Template {
	var result []Template
	for _, t := range templates {
		var finalVars []Variable
		for _, variable := range t.Variables {
			if !findInArray(varKeys, variable.Name) {
				finalVars = append(finalVars, variable)
			}
		}
		t.Variables = finalVars
		result = append(result, t)
	}
	return result
}

func findInArray(a []string, f string) bool {
	for _, s := range a {
		if s == f {
			return true
		}
	}
	return false
}
