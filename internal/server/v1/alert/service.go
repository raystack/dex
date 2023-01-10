package alert

import (
	"context"
	"strings"

	sirenv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/siren/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/odpf/dex/pkg/errors"
)

type Service struct {
	Siren sirenv1beta1.SirenServiceClient
}

func (s *Service) UpsertAlertPolicy(ctx context.Context, projectSlug string, update Policy) (*Policy, error) {
	ns, err := s.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return nil, err
	}

	alertPolicy, err := s.getAlertPolicyForResource(ctx, ns.ID, update.Resource)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, err
	}

	if alertPolicy != nil && len(alertPolicy.Rules) != 0 {
		disableRuleRequests := mapAlertPolicyToUpdateRulesRequest(*alertPolicy, ns.ID)
		for _, request := range disableRuleRequests {
			request.Enabled = false
			_, err := s.Siren.UpdateRule(ctx, request)
			if err != nil {
				return nil, err
			}
		}
	}

	updateRuleRequests := mapAlertPolicyToUpdateRulesRequest(update, ns.ID)
	for _, request := range updateRuleRequests {
		_, err := s.Siren.UpdateRule(ctx, request)
		if err != nil {
			return nil, err
		}
	}

	alertPolicy, err = s.getAlertPolicyForResource(ctx, ns.ID, update.Resource)
	if err != nil {
		return nil, err
	}

	return alertPolicy, nil
}

func (s *Service) GetAlertPolicy(ctx context.Context, projectSlug string, resource string) (*Policy, error) {
	ns, err := s.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return nil, err
	}

	alertPolicy, err := s.getAlertPolicyForResource(ctx, ns.ID, resource)
	if err != nil {
		return nil, err
	}

	return alertPolicy, nil
}

func (s *Service) ListAlerts(ctx context.Context, projectSlug string, resource string) ([]Alert, error) {
	ns, err := s.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return nil, err
	}

	alertsResp, err := s.Siren.ListAlerts(ctx, &sirenv1beta1.ListAlertsRequest{
		ProviderType: alertProviderName,
		ProviderId:   ns.Provider,
		ResourceName: resource,
	})
	if err != nil {
		return nil, err
	}

	return mapProtoAlertsToAlerts(alertsResp.GetAlerts()), nil
}

func (s *Service) ListAlertTemplates(ctx context.Context, tag string) ([]Template, error) {
	templatesResp, err := s.Siren.ListTemplates(ctx, &sirenv1beta1.ListTemplatesRequest{
		Tag: tag,
	})
	if err != nil {
		return nil, err
	}

	return mapProtoTemplatesToTemplates(templatesResp.Templates), nil
}

func (s *Service) GetAlertTemplate(ctx context.Context, urn string) (*Template, error) {
	templateResp, err := s.Siren.GetTemplate(ctx, &sirenv1beta1.GetTemplateRequest{
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

func (s *Service) getAlertPolicyForResource(ctx context.Context, providerNamespace uint64, resource string) (*Policy, error) {
	rpcReq := &sirenv1beta1.ListRulesRequest{
		Namespace:         resource,
		ProviderNamespace: providerNamespace,
	}

	rpcResp, err := s.Siren.ListRules(ctx, rpcReq)
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

func (s *Service) getNamespaceForProject(ctx context.Context, projectSlug string) (*namespace, error) {
	listNamespacesResponse, err := s.Siren.ListNamespaces(ctx, &sirenv1beta1.ListNamespacesRequest{})
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

func (s *Service) GetProjectDataSource(ctx context.Context, projectSlug string) (string, error) {
	ns, err := s.getNamespaceForProject(ctx, projectSlug)
	if err != nil {
		return "", err
	}

	return ns.Name, nil
}
