package alert

import (
	"context"
	"net/http"
	"strings"

	sirenv1beta1grpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc"
	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/pkg/errors"
)

const (
	alertPolicyNotFound      = "no Alert Policy found for given resource"
	alertProviderName        = "cortex"
	projectSlugSirenLabelKey = "projects"
)

var sirenTemplateVariables = []string{"WARN_THRESHOLD", "CRIT_THRESHOLD"}

type Service struct {
	Siren sirenv1beta1grpc.SirenServiceClient
}

func (svc *Service) HandleListTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tag := strings.TrimSpace(r.URL.Query().Get("tag"))

		templates, err := svc.ListAlertTemplates(ctx, tag)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK,
			utils.ListResponse[Template]{
				Items: RemoveSuppliedVariablesFromTemplates(templates, SuppliedVariables),
			},
		)
	}
}

func (svc *Service) UpsertAlertPolicy(ctx context.Context, projectSlug string, update Policy) (*Policy, error) {
	ns, err := svc.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return nil, err
	}

	alertPolicy, err := svc.getAlertPolicyForResource(ctx, ns.ID, update.Resource)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, err
	}

	if alertPolicy != nil && len(alertPolicy.Rules) != 0 {
		disableRuleRequests := mapAlertPolicyToUpdateRulesRequest(*alertPolicy, ns.ID)
		for _, request := range disableRuleRequests {
			request.Enabled = false
			_, err := svc.Siren.UpdateRule(ctx, request)
			if err != nil {
				return nil, err
			}
		}
	}

	updateRuleRequests := mapAlertPolicyToUpdateRulesRequest(update, ns.ID)
	for _, request := range updateRuleRequests {
		_, err := svc.Siren.UpdateRule(ctx, request)
		if err != nil {
			return nil, err
		}
	}

	alertPolicy, err = svc.getAlertPolicyForResource(ctx, ns.ID, update.Resource)
	if err != nil {
		return nil, err
	}

	return alertPolicy, nil
}

func (svc *Service) GetAlertPolicy(ctx context.Context, projectSlug, resource, resourceTag string) (*Policy, error) {
	ns, err := svc.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return nil, err
	}

	var alertPolicy *Policy
	alertPolicy, err = svc.getAlertPolicyForResource(ctx, ns.ID, resource)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, err
	}

	templates, err := svc.ListAlertTemplates(ctx, resourceTag)
	if err != nil {
		return nil, err
	}

	var rules []Rule

	for _, template := range templates {
		alertExist := false

		if alertPolicy != nil {
			for _, rule := range alertPolicy.Rules {
				if rule.Template == template.Name {
					rule.Variables = filterVariables(rule.Variables)
					rules = append(rules, rule)
					alertExist = true
					break
				}
			}
		}

		if !alertExist {
			rule := Rule{
				Variables: filterVariables(template.Variables),
				Enabled:   false,
				Template:  template.Name,
				CreatedAt: template.CreatedAt,
				UpdatedAt: template.UpdatedAt,
			}

			rules = append(rules, rule)
		}
	}

	alertPolicy = &Policy{
		Resource: resource,
		Rules:    rules,
	}

	return alertPolicy, nil
}

func (svc *Service) ListAlerts(ctx context.Context, projectSlug string, resource string) ([]Alert, error) {
	ns, err := svc.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return nil, err
	}

	alertsResp, err := svc.Siren.ListAlerts(ctx, &sirenv1beta1.ListAlertsRequest{
		ProviderType: alertProviderName,
		ProviderId:   ns.Provider,
		ResourceName: resource,
	})
	if err != nil {
		return nil, err
	}

	return mapProtoAlertsToAlerts(alertsResp.GetAlerts()), nil
}

func (svc *Service) ListAlertTemplates(ctx context.Context, tag string) ([]Template, error) {
	templatesResp, err := svc.Siren.ListTemplates(ctx, &sirenv1beta1.ListTemplatesRequest{
		Tag: tag,
	})
	if err != nil {
		return nil, err
	}

	return mapProtoTemplatesToTemplates(templatesResp.Templates), nil
}

func (svc *Service) GetAlertTemplate(ctx context.Context, urn string) (*Template, error) {
	templateResp, err := svc.Siren.GetTemplate(ctx, &sirenv1beta1.GetTemplateRequest{
		Name: urn,
	})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errors.ErrNotFound.
				WithMsgf("no Alert Template found with given name").
				WithCausef(st.Message())
		}
		return nil, err
	}

	resp := mapProtoTemplateToTemplate(templateResp.Template)
	return &resp, nil
}

func (svc *Service) getAlertPolicyForResource(ctx context.Context, providerNamespace uint64, resource string) (*Policy, error) {
	rpcReq := &sirenv1beta1.ListRulesRequest{
		Namespace:         resource,
		ProviderNamespace: providerNamespace,
	}

	rpcResp, err := svc.Siren.ListRules(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	alertPolicies := mapRulesToAlertPolicy(rpcResp.Rules)
	if len(alertPolicies) > 1 {
		return nil, errors.ErrInternal.
			WithMsgf("more than 1 Alert policies for a resource").
			WithCausef("bad upstream response")
	} else if len(alertPolicies) == 0 {
		return nil, errors.ErrNotFound.WithMsgf(alertPolicyNotFound)
	}

	return &alertPolicies[0], nil
}

func (svc *Service) getNamespaceForProject(ctx context.Context, projectSlug string) (*namespace, error) {
	listNamespacesResponse, err := svc.Siren.ListNamespaces(ctx, &sirenv1beta1.ListNamespacesRequest{})
	if err != nil {
		return nil, err
	}
	for _, namespace := range listNamespacesResponse.GetNamespaces() {
		projects := strings.Split(namespace.Labels[projectSlugSirenLabelKey], ",")
		for _, project := range projects {
			if project == projectSlug {
				return mapProtoNamespaceToNamespace(namespace), nil
			}
		}
	}
	return nil, errors.ErrNotFound.WithMsgf("Alert namespace not found for given project id")
}

func (svc *Service) GetProjectDataSource(ctx context.Context, projectSlug string) (string, error) {
	ns, err := svc.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return "", err
	}

	return ns.Name, nil
}

func filterVariables(variables []Variable) []Variable {
	var allowedVariables []Variable
	for _, variable := range variables {
		for _, sirenTemplateVariable := range sirenTemplateVariables {
			if variable.Name == sirenTemplateVariable {
				allowedVariables = append(allowedVariables, variable)
			}
		}
	}

	return allowedVariables
}
