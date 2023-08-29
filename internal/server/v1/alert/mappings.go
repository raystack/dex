package alert

import (
	"fmt"
	"strconv"
	"time"

	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"
	sirenReceiverPkg "github.com/goto/siren/core/receiver"

	"github.com/goto/dex/generated/models"
)

const defaultDecimalBase = 10

type Variable struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Default     string `json:"default,omitempty"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Rule struct {
	ID        string     `json:"id,omitempty"`
	Template  string     `json:"template"`
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Variables []Variable `json:"variables"`
}

type Policy struct {
	Resource string `json:"resource"`
	Rules    []Rule `json:"rules"`
}

type Alert struct {
	ID          string    `json:"id"`
	Resource    string    `json:"resource"`
	Metric      string    `json:"metric"`
	Value       string    `json:"value"`
	Severity    string    `json:"severity"`
	Rule        string    `json:"Rule"`
	TriggeredAt time.Time `json:"triggered_at"`
}

type Template struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Body      string     `json:"body"`
	Tags      []string   `json:"tags"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Variables []Variable `json:"variables"`
}

type namespace struct {
	ID          uint64                 `json:"id"`
	URN         string                 `json:"urn"`
	Name        string                 `json:"name"`
	Provider    uint64                 `json:"provider"`
	Credentials map[string]interface{} `json:"credentials"`
	Labels      map[string]string      `json:"labels"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func mapRulesToAlertPolicy(rules []*sirenv1beta1.Rule) []Policy {
	resourceRules := map[string][]Rule{}
	for _, rule := range rules {
		resource, rule := mapProtoRuleToRule(rule)
		resourceRules[resource] = append(resourceRules[resource], rule)
	}

	var policies []Policy
	for resource, rules := range resourceRules {
		policies = append(policies, Policy{
			Resource: resource,
			Rules:    rules,
		})
	}
	return policies
}

func mapAlertPolicyToUpdateRulesRequest(p Policy, providerNamespace uint64) []*sirenv1beta1.UpdateRuleRequest {
	var rules []*sirenv1beta1.UpdateRuleRequest
	for _, r := range p.Rules {
		rules = append(rules, &sirenv1beta1.UpdateRuleRequest{
			GroupName:         r.Template,
			Namespace:         p.Resource,
			Template:          r.Template,
			ProviderNamespace: providerNamespace,
			Enabled:           r.Enabled,
			Variables:         mapVariablesToProtoRuleVariables(r.Variables),
		})
	}
	return rules
}

func mapProtoRuleToRule(r *sirenv1beta1.Rule) (string, Rule) {
	return r.Namespace, Rule{
		ID:        strconv.FormatUint(r.GetId(), defaultDecimalBase),
		Template:  r.GetTemplate(),
		Enabled:   r.Enabled,
		CreatedAt: r.GetCreatedAt().AsTime(),
		UpdatedAt: r.GetUpdatedAt().AsTime(),
		Variables: mapProtoRuleVariablesToVariables(r.GetVariables()),
	}
}

func mapProtoRuleVariablesToVariables(variables []*sirenv1beta1.Variables) []Variable {
	var result []Variable
	for _, v := range variables {
		result = append(result, Variable{
			Name:        v.GetName(),
			Value:       v.GetValue(),
			Type:        v.GetType(),
			Description: v.GetDescription(),
		})
	}
	return result
}

func mapVariablesToProtoRuleVariables(variables []Variable) []*sirenv1beta1.Variables {
	var result []*sirenv1beta1.Variables
	for _, v := range variables {
		result = append(result, &sirenv1beta1.Variables{
			Name:        v.Name,
			Value:       v.Value,
			Type:        v.Type,
			Description: v.Description,
		})
	}
	return result
}

func mapProtoAlertsToAlerts(alerts []*sirenv1beta1.Alert) []Alert {
	var result []Alert
	for _, a := range alerts {
		result = append(result, Alert{
			ID:          strconv.FormatUint(a.Id, defaultDecimalBase),
			Resource:    a.ResourceName,
			Metric:      a.MetricName,
			Value:       a.MetricValue,
			Severity:    a.Severity,
			Rule:        a.Rule,
			TriggeredAt: a.TriggeredAt.AsTime(),
		})
	}
	return result
}

func mapProtoTemplateToTemplate(t *sirenv1beta1.Template) Template {
	return Template{
		ID:        strconv.FormatUint(t.GetId(), defaultDecimalBase),
		Name:      t.GetName(),
		Body:      t.GetBody(),
		Tags:      t.Tags,
		CreatedAt: t.GetCreatedAt().AsTime(),
		UpdatedAt: t.GetUpdatedAt().AsTime(),
		Variables: mapProtoTemplateVariablesToVariables(t.GetVariables()),
	}
}

func mapProtoTemplatesToTemplates(templates []*sirenv1beta1.Template) []Template {
	var result []Template
	for _, t := range templates {
		result = append(result, mapProtoTemplateToTemplate(t))
	}
	return result
}

func mapProtoTemplateVariablesToVariables(variables []*sirenv1beta1.TemplateVariables) []Variable {
	var result []Variable
	for _, v := range variables {
		result = append(result, Variable{
			Name:        v.GetName(),
			Default:     v.GetDefault(),
			Type:        v.GetType(),
			Description: v.GetDescription(),
		})
	}
	return result
}

func mapProtoNamespaceToNamespace(ns *sirenv1beta1.Namespace) *namespace {
	return &namespace{
		ID:          ns.GetId(),
		URN:         ns.GetUrn(),
		Name:        ns.GetName(),
		Provider:    ns.GetProvider(),
		Credentials: ns.GetCredentials().AsMap(),
		Labels:      ns.GetLabels(),
		CreatedAt:   ns.GetCreatedAt().AsTime(),
		UpdatedAt:   ns.GetUpdatedAt().AsTime(),
	}
}

func mapSirenReceiverToAlertChannel(recv *sirenv1beta1.Receiver) models.AlertChannel {
	channelCriticality := recv.Labels[sirenReceiverLabelKeySeverity]

	configMap := recv.Configurations.AsMap()
	var channelName string
	channelNameAny, exists := configMap[sirenReceiverConfigKeyChannelName]
	if exists {
		channelName = fmt.Sprintf("%v", channelNameAny)
	}

	var serviceKey string
	serviceKeyAny, exists := configMap[sirenReceiverConfigKeyServiceKey]
	if exists {
		serviceKey = fmt.Sprintf("%v", serviceKeyAny)
	}

	return models.AlertChannel{
		ReceiverID:          fmt.Sprintf("%d", recv.Id),
		ReceiverName:        recv.Name,
		ChannelName:         channelName,
		ChannelCriticality:  models.NewChannelCriticality(models.ChannelCriticality(channelCriticality)),
		ChannelType:         mapReceiverTypeToChannelType(recv.Type),
		PagerdutyServiceKey: serviceKey,
	}
}

func mapReceiverTypeToChannelType(receiverType string) *models.AlertChannelType {
	if receiverType == "" {
		return nil
	}

	var val models.AlertChannelType
	switch receiverType {
	case sirenReceiverPkg.TypePagerDuty:
		val = models.AlertChannelTypePagerduty
	case sirenReceiverPkg.TypeSlackChannel:
		val = models.AlertChannelTypeSlackChannel
	}

	return models.NewAlertChannelType(val)
}

func isValidReceiverType(val string) bool {
	switch val {
	case sirenReceiverPkg.TypePagerDuty,
		sirenReceiverPkg.TypeSlackChannel:
		return true
	default:
		return false
	}
}
