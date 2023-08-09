package alert

import (
	"context"
	"fmt"

	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	sirenv1beta1grpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"
	sirenReceiverPkg "github.com/goto/siren/core/receiver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
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

	metadata, err := buildSubscriptionMetadataMap(form, project.Slug, group.Slug)
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

	metadata, err := buildSubscriptionMetadataMap(form, project.Slug, group.Slug)
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

func (svc *SubscriptionService) fetchShieldData(
	ctx context.Context,
	projectID, groupID string,
) (*shieldv1beta1.Project, *shieldv1beta1.Group, uint64, error) {
	project, err := svc.getProject(ctx, projectID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error getting shield's project: %w", err)
	}
	namespaceID, err := svc.getSirenNamespaceID(project)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error getting siren namespace: %w", err)
	}

	group, err := svc.getGroup(ctx, groupID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error getting shield's group: %w", err)
	}

	return project, group, namespaceID, nil
}

func (svc *SubscriptionService) getGroup(ctx context.Context, groupID string) (*shieldv1beta1.Group, error) {
	resp, err := svc.shieldClient.GetGroup(ctx, &shieldv1beta1.GetGroupRequest{
		Id: groupID,
	})
	if err != nil {
		return nil, err
	}

	return resp.Group, nil
}

func (svc *SubscriptionService) getProject(ctx context.Context, projectID string) (*shieldv1beta1.Project, error) {
	resp, err := svc.shieldClient.GetProject(ctx, &shieldv1beta1.GetProjectRequest{
		Id: projectID,
	})
	if err != nil {
		return nil, err
	}

	return resp.Project, nil
}

func (*SubscriptionService) getSirenNamespaceID(project *shieldv1beta1.Project) (uint64, error) {
	projectMetadata := project.GetMetadata().AsMap()

	namespaceIDAny, exists := projectMetadata["siren_namespace"]
	if !exists {
		return 0, ErrNoShieldSirenNamespace
	}
	namespaceID, ok := namespaceIDAny.(float64)
	if !ok {
		return 0, ErrNoShieldSirenNamespace
	}

	return uint64(namespaceID), nil
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

func buildSubscriptionMetadataMap(form SubscriptionForm, projectSlug, groupSlug string) (*structpb.Struct, error) {
	metadata, err := structpb.NewStruct(map[string]interface{}{
		"group_id":            form.GroupID,
		"resource_type":       form.ResourceType,
		"resource_id":         form.ResourceID,
		"project_id":          form.ProjectID,
		"group_slug":          groupSlug,
		"project_slug":        projectSlug,
		"channel_criticality": string(form.ChannelCriticality),
	})
	if err != nil {
		return nil, fmt.Errorf("error building metadata: %w", err)
	}

	return metadata, nil
}

func buildSubscriptionURN(form SubscriptionForm, groupSlug string) string {
	return fmt.Sprintf("%s:%s:%s:%s", groupSlug, form.AlertSeverity, form.ResourceType, form.ResourceID)
}
