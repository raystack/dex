package alert

import (
	"context"
	"fmt"
	"strings"

	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	sirenv1beta1grpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"
	sirenReceiverPkg "github.com/goto/siren/core/receiver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/goto/dex/generated/models"
)

type SubscriptionService struct {
	sirenClient  sirenv1beta1grpc.SirenServiceClient
	shieldClient shieldv1beta1rpc.ShieldServiceClient
}

func NewSubscriptionService(
	sirenClient sirenv1beta1grpc.SirenServiceClient,
	shieldClient shieldv1beta1rpc.ShieldServiceClient,
) *SubscriptionService {
	return &SubscriptionService{
		sirenClient:  sirenClient,
		shieldClient: shieldClient,
	}
}

func (svc *SubscriptionService) FindSubscription(ctx context.Context, subscriptionID int) (*sirenv1beta1.Subscription, error) {
	request := &sirenv1beta1.GetSubscriptionRequest{
		Id: uint64(subscriptionID),
	}

	resp, err := svc.sirenClient.GetSubscription(ctx, request)
	if err != nil {
		stat := status.Convert(err)
		if stat.Code() == codes.NotFound {
			return nil, ErrSubscriptionNotFound
		}

		return nil, err
	}

	return resp.Subscription, nil
}

func (svc *SubscriptionService) GetSubscriptions(ctx context.Context, groupID, resourceID, resourceType string) ([]*sirenv1beta1.Subscription, error) {
	request := &sirenv1beta1.ListSubscriptionsRequest{
		Metadata: make(map[string]string),
	}
	if groupID != "" {
		request.Metadata[groupMetadataKey] = groupID
	}
	if resourceID != "" {
		request.Metadata[resourceIDMetadataKey] = resourceID
	}
	if resourceType != "" {
		request.Metadata[resourceTypeMetadataKey] = resourceType
	}

	resp, err := svc.sirenClient.ListSubscriptions(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp.Subscriptions, nil
}

func (svc *SubscriptionService) CreateSubscription(ctx context.Context, form SubscriptionForm) (int, error) {
	project, group, namespaceID, err := svc.fetchShieldData(ctx, form.ProjectID, form.GroupID)
	if err != nil {
		return 0, err
	}
	receiver, err := svc.getSirenReceiver(ctx, group.Slug, form.ChannelCriticality)
	if err != nil {
		return 0, fmt.Errorf("error getting siren's receiver: %w", err)
	}

	channelName := svc.getSirenReceiverConfigValues(receiver, "channel_name")[0]
	metadata, err := buildSubscriptionMetadataMap(form, project.Slug, group.Slug, channelName)
	if err != nil {
		return 0, err
	}

	request := &sirenv1beta1.CreateSubscriptionRequest{
		Urn:       buildSubscriptionURN(form, group.GetSlug()),
		Namespace: namespaceID,
		Receivers: []*sirenv1beta1.ReceiverMetadata{
			{
				Id: receiver.Id,
			},
		},
		Match: map[string]string{
			"severity":   string(form.AlertSeverity),
			"identifier": form.ResourceID,
		},
		Metadata:  metadata,
		CreatedBy: form.UserID,
	}

	resp, err := svc.sirenClient.CreateSubscription(ctx, request)
	if err != nil {
		return 0, err
	}

	return int(resp.Id), nil
}

func (svc *SubscriptionService) UpdateSubscription(ctx context.Context, subscriptionID int, form SubscriptionForm) error {
	project, group, namespaceID, err := svc.fetchShieldData(ctx, form.ProjectID, form.GroupID)
	if err != nil {
		return err
	}
	receiver, err := svc.getSirenReceiver(ctx, group.Slug, form.ChannelCriticality)
	if err != nil {
		return fmt.Errorf("error getting siren's receiver: %w", err)
	}

	channelName := svc.getSirenReceiverConfigValues(receiver, "channel_name")[0]
	metadata, err := buildSubscriptionMetadataMap(form, project.Slug, group.Slug, channelName)
	if err != nil {
		return err
	}
	request := &sirenv1beta1.UpdateSubscriptionRequest{
		Id:        uint64(subscriptionID),
		Urn:       buildSubscriptionURN(form, group.GetSlug()),
		Namespace: namespaceID,
		Receivers: []*sirenv1beta1.ReceiverMetadata{
			{Id: receiver.Id},
		},
		Match: map[string]string{
			"severity":   string(form.AlertSeverity),
			"identifier": form.ResourceID,
		},
		Metadata:  metadata,
		UpdatedBy: form.UserID,
	}

	_, err = svc.sirenClient.UpdateSubscription(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (svc *SubscriptionService) DeleteSubscription(ctx context.Context, subscriptionID int) error {
	request := &sirenv1beta1.DeleteSubscriptionRequest{
		Id: uint64(subscriptionID),
	}
	_, err := svc.sirenClient.DeleteSubscription(ctx, request)
	if err != nil {
		stat := status.Convert(err)
		if stat.Code() == codes.NotFound {
			return ErrSubscriptionNotFound
		}

		return err
	}

	return nil
}

func (svc *SubscriptionService) GetAlertChannels(ctx context.Context, groupID string) ([]models.AlertChannel, error) {
	group, err := svc.getGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("error getting group: %w", err)
	}

	receivers, err := svc.getGroupSirenReceivers(ctx, group.Slug)
	if err != nil {
		return nil, fmt.Errorf("error getting receivers: %w", err)
	}

	alertChannels := []models.AlertChannel{}
	for _, recv := range receivers {
		severity := svc.getSirenReceiverLabelValues(recv, sirenReceiverLabelKeySeverity)[0]
		configs := svc.getSirenReceiverConfigValues(recv, sirenReceiverConfigKeyChannelName, sirenReceiverConfigKeyServiceKey)
		channelName := configs[0]
		pagerdutyServiceKey := configs[1]

		var channelType models.AlertChannelType
		if channelName != "" {
			channelType = models.AlertChannelTypeSlackChannel
		} else if pagerdutyServiceKey != "" {
			channelType = models.AlertChannelTypePagerduty
		}

		alertChannels = append(alertChannels, models.AlertChannel{
			ReceiverID:          fmt.Sprint(recv.Id),
			ReceiverName:        recv.Name,
			ChannelName:         channelName,
			PagerdutyServiceKey: pagerdutyServiceKey,
			ChannelCriticality:  models.NewChannelCriticality(models.ChannelCriticality(severity)),
			ChannelType:         models.NewAlertChannelType(channelType),
		})
	}

	return alertChannels, nil
}

func (svc *SubscriptionService) SetAlertChannels(ctx context.Context, userID string, groupID string, forms []AlertChannelForm) ([]models.AlertChannel, error) {
	group, err := svc.getGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("error getting group: %w", err)
	}
	org, err := svc.getOrg(ctx, group.GetOrgId())
	if err != nil {
		return nil, fmt.Errorf("error getting group: %w", err)
	}

	parentReceiver, namespaceID, err := svc.extractShieldOrgMetadata(org)
	if err != nil {
		return nil, fmt.Errorf("error getting parent slack receivers: %w", err)
	}
	receivers, err := svc.getGroupSirenReceivers(ctx, group.GetSlug())
	if err != nil {
		return nil, fmt.Errorf("error getting receivers: %w", err)
	}

	results := make([]models.AlertChannel, len(forms))
	for i, form := range forms {
		if !isValidReceiverType(form.ChannelType) {
			continue // skip
		}

		currentReceiver := receivers.Find(form.ChannelType, string(form.ChannelCriticality))
		if currentReceiver != nil { // update flow
			err = svc.updateReceiver(ctx, currentReceiver, form)
			if err != nil {
				return nil, fmt.Errorf("error updating receiver (index=\"%d\"): %w", i, err)
			}
		} else { // create flow
			newReceiver, err := svc.createReceiver(
				ctx,
				form,
				parentReceiver, // used for slack_channel type only
				group.GetSlug(),
				org.GetSlug(),
			)
			if err != nil {
				return nil, fmt.Errorf("error creating receiver (index=\"%d\"): %w", i, err)
			}

			// create a subscription using the new receiver
			_, err = svc.sirenClient.CreateSubscription(ctx, &sirenv1beta1.CreateSubscriptionRequest{
				Urn:       fmt.Sprintf("%s-%s-%s", org.GetSlug(), group.GetSlug(), strings.ToLower(string(form.ChannelCriticality))),
				Namespace: namespaceID,
				Receivers: []*sirenv1beta1.ReceiverMetadata{
					{
						Id: newReceiver.Id,
					},
				},
				Match: map[string]string{
					"severity": string(form.ChannelCriticality),
					"team":     group.GetSlug(),
				},
				CreatedBy: userID,
			})
			if err != nil {
				return nil, fmt.Errorf("error creating subscription: %w", err)
			}

			currentReceiver = newReceiver
		}

		results[i] = mapSirenReceiverToAlertChannel(currentReceiver)
	}

	return results, nil
}

func (svc *SubscriptionService) createReceiver(ctx context.Context, form AlertChannelForm, slackReceiverParentID uint64, groupSlug, orgSlug string) (*sirenv1beta1.Receiver, error) {
	receiverType := form.ChannelType

	var parentID uint64
	configMap := map[string]interface{}{}
	if receiverType == sirenReceiverPkg.TypeSlackChannel {
		parentID = slackReceiverParentID
		configMap[sirenReceiverConfigKeyChannelName] = form.ChannelName
	} else if receiverType == sirenReceiverPkg.TypePagerDuty {
		configMap[sirenReceiverConfigKeyServiceKey] = form.PagerdutyServiceKey
	}

	receiverConfig, err := structpb.NewStruct(configMap)
	if err != nil {
		return nil, fmt.Errorf("error building receiver configuration: %w", err)
	}
	receiver := &sirenv1beta1.Receiver{
		Name: fmt.Sprintf("%s-%s-%s-%s", orgSlug, groupSlug, form.ChannelType, strings.ToLower(string(form.ChannelCriticality))),
		Type: receiverType,
		Labels: map[string]string{
			sirenReceiverLabelKeyOrg:      orgSlug,
			sirenReceiverLabelKeyTeam:     groupSlug,
			sirenReceiverLabelKeySeverity: string(form.ChannelCriticality),
		},
		ParentId:       parentID,
		Configurations: receiverConfig,
	}
	resp, err := svc.sirenClient.CreateReceiver(ctx, &sirenv1beta1.CreateReceiverRequest{
		Name:           receiver.Name,
		Type:           receiver.Type,
		ParentId:       receiver.ParentId,
		Labels:         receiver.Labels,
		Configurations: receiver.Configurations,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating receiver: %w", err)
	}

	receiver.Id = resp.GetId()
	return receiver, nil
}

func (svc *SubscriptionService) updateReceiver(ctx context.Context, receiver *sirenv1beta1.Receiver, form AlertChannelForm) error {
	receiverType := form.ChannelType
	configMap := receiver.Configurations.AsMap()

	if receiverType == sirenReceiverPkg.TypeSlackChannel {
		configMap[sirenReceiverConfigKeyChannelName] = form.ChannelName
	} else if receiverType == sirenReceiverPkg.TypePagerDuty {
		configMap[sirenReceiverConfigKeyServiceKey] = form.PagerdutyServiceKey
	}

	newConfig, err := structpb.NewStruct(configMap)
	if err != nil {
		return fmt.Errorf("error building new receiver configuration: %w", err)
	}
	receiver.Configurations = newConfig

	_, err = svc.sirenClient.UpdateReceiver(ctx, &sirenv1beta1.UpdateReceiverRequest{
		Id:             receiver.Id,
		Name:           receiver.Name,
		ParentId:       receiver.ParentId,
		Labels:         receiver.Labels,
		Configurations: receiver.Configurations,
	})
	if err != nil {
		return fmt.Errorf("error updating receiver: %w", err)
	}

	return nil
}

func (svc *SubscriptionService) fetchShieldData(
	ctx context.Context,
	projectID, groupID string,
) (*shieldv1beta1.Project, *shieldv1beta1.Group, uint64, error) {
	project, err := svc.getProject(ctx, projectID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error getting shield's project: %w", err)
	}

	org, err := svc.getOrg(ctx, project.OrgId)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error getting shield's org: %w", err)
	}
	namespaceIDAny, exists := org.Metadata.AsMap()[shieldOrgMetadataKeySirenNamespaceID]
	if !exists {
		return nil, nil, 0, ErrNoShieldSirenNamespace
	}
	namespaceIDFloat, validNamespace := namespaceIDAny.(float64)
	if !validNamespace {
		return nil, nil, 0, ErrInvalidShieldSirenNamespace
	}

	group, err := svc.getGroup(ctx, groupID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error getting shield's group: %w", err)
	}

	return project, group, uint64(namespaceIDFloat), nil
}

func (svc *SubscriptionService) getGroup(ctx context.Context, groupID string) (*shieldv1beta1.Group, error) {
	resp, err := svc.shieldClient.GetGroup(ctx, &shieldv1beta1.GetGroupRequest{
		Id: groupID,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, ErrNoShieldGroup
			}
		}
		return nil, err
	}

	return resp.Group, nil
}

func (svc *SubscriptionService) getOrg(ctx context.Context, orgID string) (*shieldv1beta1.Organization, error) {
	resp, err := svc.shieldClient.GetOrganization(ctx, &shieldv1beta1.GetOrganizationRequest{
		Id: orgID,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, ErrNoShieldOrg
			}
		}
		return nil, err
	}

	return resp.Organization, nil
}

func (svc *SubscriptionService) getProject(ctx context.Context, projectID string) (*shieldv1beta1.Project, error) {
	resp, err := svc.shieldClient.GetProject(ctx, &shieldv1beta1.GetProjectRequest{
		Id: projectID,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, ErrNoShieldProject
			}
		}
		return nil, err
	}

	return resp.Project, nil
}

func (svc *SubscriptionService) getGroupSirenReceivers(ctx context.Context, groupSlug string) (SirenReceivers, error) {
	resp, err := svc.sirenClient.ListReceivers(ctx, &sirenv1beta1.ListReceiversRequest{
		Labels: map[string]string{
			sirenReceiverLabelKeyTeam: groupSlug,
		},
	})
	if err != nil {
		return nil, err
	}

	return resp.Receivers, nil
}

func (*SubscriptionService) extractShieldOrgMetadata(org *shieldv1beta1.Organization) (uint64, uint64, error) {
	metadata := org.Metadata.AsMap()
	receiverIDAny, exists := metadata[shieldOrgMetadataKeySirenParentSlackReceiverID]
	if !exists {
		return 0, 0, ErrNoShieldParentSlackReceiver
	}
	receiverIDFloat, validReceiver := receiverIDAny.(float64)
	if !validReceiver {
		return 0, 0, ErrInvalidShieldParentSlackReceiver
	}

	namespaceIDAny, exists := metadata[shieldOrgMetadataKeySirenNamespaceID]
	if !exists {
		return 0, 0, ErrNoShieldSirenNamespace
	}
	namespaceIDFloat, validNamespace := namespaceIDAny.(float64)
	if !validNamespace {
		return 0, 0, ErrInvalidShieldSirenNamespace
	}

	return uint64(receiverIDFloat), uint64(namespaceIDFloat), nil
}

func (svc *SubscriptionService) getSirenReceiver(ctx context.Context, groupSlug string, criticality ChannelCriticality) (*sirenv1beta1.Receiver, error) {
	resp, err := svc.sirenClient.ListReceivers(ctx, &sirenv1beta1.ListReceiversRequest{
		Labels: map[string]string{
			"team":     groupSlug,
			"severity": string(criticality),
		},
	})
	if err != nil {
		return nil, err
	}
	receivers := resp.Receivers
	if len(receivers) == 0 {
		return nil, ErrNoSirenReceiver
	}
	if len(receivers) > 1 {
		return nil, ErrMultipleSirenReceiver
	}
	receiver := receivers[0]
	if receiver.Type != sirenReceiverPkg.TypeSlackChannel {
		return nil, ErrInvalidSirenReceiver
	}

	return receiver, nil
}

func (*SubscriptionService) getSirenReceiverConfigValues(receiver *sirenv1beta1.Receiver, keys ...string) []string {
	configMap := receiver.Configurations.AsMap()

	values := make([]string, len(keys))
	for i, key := range keys {
		valueAny, exists := configMap[key]
		if exists {
			value, ok := valueAny.(string)
			if ok {
				values[i] = value
			}
		}
	}

	return values
}

func (*SubscriptionService) getSirenReceiverLabelValues(receiver *sirenv1beta1.Receiver, keys ...string) []string {
	labels := receiver.GetLabels()

	values := make([]string, len(keys))
	for i, key := range keys {
		value, exists := labels[key]
		if exists {
			values[i] = value
		}
	}

	return values
}

func buildSubscriptionMetadataMap(form SubscriptionForm, projectSlug, groupSlug, channelName string) (*structpb.Struct, error) {
	metadata, err := structpb.NewStruct(map[string]interface{}{
		"group_id":            form.GroupID,
		"resource_type":       form.ResourceType,
		"resource_id":         form.ResourceID,
		"project_id":          form.ProjectID,
		"group_slug":          groupSlug,
		"project_slug":        projectSlug,
		"channel_criticality": string(form.ChannelCriticality),
		"channel_name":        channelName,
	})
	if err != nil {
		return nil, fmt.Errorf("error building metadata: %w", err)
	}

	return metadata, nil
}

func buildSubscriptionURN(form SubscriptionForm, groupSlug string) string {
	return fmt.Sprintf("%s:%s:%s:%s", groupSlug, form.AlertSeverity, form.ResourceType, form.ResourceID)
}
