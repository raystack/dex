package alert_test

import (
	"context"
	"fmt"
	"testing"

	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"
	sirenReceiverPkg "github.com/goto/siren/core/receiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/v1/alert"
	"github.com/goto/dex/mocks"
)

func TestSubscriptionServiceFindSubscription(t *testing.T) {
	ctx := context.TODO()
	subscriptionID := 105

	t.Run("should return subscription on success", func(t *testing.T) {
		subscription := &sirenv1beta1.Subscription{
			Id:        uint64(subscriptionID),
			Urn:       "sample-urn",
			Namespace: 1,
			Receivers: []*sirenv1beta1.ReceiverMetadata{
				{Id: 30},
			},
		}

		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("GetSubscription", ctx, &sirenv1beta1.GetSubscriptionRequest{Id: subscription.Id}).
			Return(&sirenv1beta1.GetSubscriptionResponse{
				Subscription: subscription,
			}, nil)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		result, err := service.FindSubscription(ctx, subscriptionID)
		assert.NoError(t, err)
		assert.Equal(t, subscription, result)
	})

	t.Run("should return not found error if optimus return NotFound code", func(t *testing.T) {
		grpcError := status.Error(codes.NotFound, "Not Found")

		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("GetSubscription", ctx, &sirenv1beta1.GetSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, grpcError)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		_, err := service.FindSubscription(ctx, subscriptionID)
		assert.ErrorIs(t, err, alert.ErrSubscriptionNotFound)
	})

	t.Run("should return if client return error", func(t *testing.T) {
		expectedError := status.Error(codes.Internal, "Internal")

		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("GetSubscription", ctx, &sirenv1beta1.GetSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, expectedError)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		_, err := service.FindSubscription(ctx, subscriptionID)
		assert.ErrorIs(t, err, expectedError)
	})
}

func TestSubscriptionServiceGetSubscriptions(t *testing.T) {
	ctx := context.TODO()

	t.Run("should return subscription on success", func(t *testing.T) {
		groupID := "19293012i31"
		resourceID := "sample-resource-id-or-urn"
		resourceType := "firehose"

		subscriptions := []*sirenv1beta1.Subscription{
			{
				Id:        1,
				Urn:       "sample-urn-1",
				Namespace: 1,
				Receivers: []*sirenv1beta1.ReceiverMetadata{
					{Id: 30},
				},
			},
			{
				Id:        2,
				Urn:       "sample-urn-2",
				Namespace: 2,
				Receivers: []*sirenv1beta1.ReceiverMetadata{
					{Id: 33},
				},
			},
		}

		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("ListSubscriptions", ctx, &sirenv1beta1.ListSubscriptionsRequest{Metadata: map[string]string{
			"group_id":      groupID,
			"resource_id":   resourceID,
			"resource_type": resourceType,
		}}).
			Return(&sirenv1beta1.ListSubscriptionsResponse{
				Subscriptions: subscriptions,
			}, nil)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		result, err := service.GetSubscriptions(ctx, groupID, resourceID, resourceType)
		assert.NoError(t, err)
		assert.Equal(t, subscriptions, result)
	})

	t.Run("should return if client return error", func(t *testing.T) {
		expectedError := status.Error(codes.Internal, "Internal")

		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("ListSubscriptions", ctx, &sirenv1beta1.ListSubscriptionsRequest{Metadata: map[string]string{}}).
			Return(nil, expectedError)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		_, err := service.GetSubscriptions(ctx, "", "", "")
		assert.ErrorIs(t, err, expectedError)
	})
}

func TestSubscriptionServiceCreateSubscription(t *testing.T) {
	var (
		ctx       = context.TODO()
		groupID   = "8a7219cd-53c9-47f1-9387-5cac7abe4dcb"
		projectID = "5dab4194-9516-421a-aafe-72fd3d96ec56"
	)

	t.Run("should return error if siren namespace cannot be retrieved from project", func(t *testing.T) {
		tests := []struct {
			name     string
			metadata *structpb.Struct
		}{
			{
				name:     "empty metadata",
				metadata: nil,
			},
			{
				name: "empty metadata.siren_namespace",
				metadata: newStruct(t, map[string]interface{}{
					"siren_namespace": nil,
				}),
			},
			{
				name: "invalid format for metadata.siren_namespace",
				metadata: newStruct(t, map[string]interface{}{
					"siren_namespace": "wrong-format",
				}),
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				form := alert.SubscriptionForm{
					ProjectID: projectID,
					GroupID:   groupID,
				}
				shieldProject := &shieldv1beta1.Project{
					Slug:     "test-project",
					Metadata: test.metadata,
				}

				shield := new(mocks.ShieldServiceClient)
				shield.On("GetProject", ctx, &shieldv1beta1.GetProjectRequest{Id: projectID}).
					Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
				defer shield.AssertExpectations(t)
				client := new(mocks.SirenServiceClient)
				defer client.AssertExpectations(t)

				service := alert.NewSubscriptionService(client, shield)
				_, err := service.CreateSubscription(ctx, form)
				assert.ErrorIs(t, err, alert.ErrNoShieldSirenNamespace)
			})
		}
	})

	t.Run("should return error on failing to get siren's receiver", func(t *testing.T) {
		tests := []struct {
			name          string
			receivers     []*sirenv1beta1.Receiver
			expectedError error
		}{
			{
				name:          "nil receivers",
				receivers:     nil,
				expectedError: alert.ErrNoSirenReceiver,
			},
			{
				name:          "empty receivers",
				receivers:     []*sirenv1beta1.Receiver{},
				expectedError: alert.ErrNoSirenReceiver,
			},
			{
				name: "more than one receivers",
				receivers: []*sirenv1beta1.Receiver{
					{Id: 1},
					{Id: 2},
				},
			},
			{
				name: "receiver is not slack_channel type",
				receivers: []*sirenv1beta1.Receiver{
					{Id: 1, Type: "invalid-type"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				form := alert.SubscriptionForm{
					ProjectID:          projectID,
					GroupID:            groupID,
					ChannelCriticality: alert.ChannelCriticalityInfo,
				}
				shieldGroup := &shieldv1beta1.Group{
					Slug: "test-group",
				}
				shieldProject := &shieldv1beta1.Project{
					Slug: "test-project",
					Metadata: newStruct(t, map[string]interface{}{
						"siren_namespace": 5,
					}),
				}

				shield := new(mocks.ShieldServiceClient)
				shield.On("GetProject", ctx, &shieldv1beta1.GetProjectRequest{Id: projectID}).
					Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
				shield.On("GetGroup", ctx, &shieldv1beta1.GetGroupRequest{Id: form.GroupID}).
					Return(&shieldv1beta1.GetGroupResponse{Group: shieldGroup}, nil)
				defer shield.AssertExpectations(t)
				siren := new(mocks.SirenServiceClient)
				siren.On("ListReceivers", ctx, &sirenv1beta1.ListReceiversRequest{
					Labels: map[string]string{
						"team":     shieldGroup.Slug,
						"severity": string(form.ChannelCriticality),
					},
				}).Return(&sirenv1beta1.ListReceiversResponse{
					Receivers: test.receivers,
				}, nil)
				defer siren.AssertExpectations(t)

				service := alert.NewSubscriptionService(siren, shield)
				_, err := service.CreateSubscription(ctx, form)
				if test.expectedError != nil {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("should create subscription on success", func(t *testing.T) {
		receiverID := uint64(15)
		sirenNamespace := 5
		channelName := "test-alert-channel"

		// inputs
		form := alert.SubscriptionForm{
			UserID:             "john.doe@example.com",
			AlertSeverity:      alert.AlertSeverityCritical,
			ChannelCriticality: alert.ChannelCriticalityInfo,
			GroupID:            groupID,
			ProjectID:          projectID,
			ResourceType:       "firehose",
			ResourceID:         "test-job",
		}

		// conditions
		shieldGroup := &shieldv1beta1.Group{
			Slug: "test-group",
		}
		shieldProject := &shieldv1beta1.Project{
			Slug: "my-project-1",
			Metadata: newStruct(t, map[string]interface{}{
				"siren_namespace": sirenNamespace,
			}),
		}
		sirenReceivers := []*sirenv1beta1.Receiver{
			{Id: receiverID, Type: sirenReceiverPkg.TypeSlackChannel, Configurations: newStruct(t, map[string]interface{}{
				"channel_name": channelName,
			})},
		}

		// expectations
		expectedSirenPayload := &sirenv1beta1.CreateSubscriptionRequest{
			Urn: fmt.Sprintf(
				"%s:%s:%s:%s",
				shieldGroup.GetSlug(), form.AlertSeverity, form.ResourceType, form.ResourceID,
			),
			Namespace: uint64(sirenNamespace),
			Receivers: []*sirenv1beta1.ReceiverMetadata{
				{Id: receiverID},
			},
			Match: map[string]string{
				"severity":   string(alert.AlertSeverityCritical),
				"identifier": "test-job",
			},
			Metadata: newStruct(t, map[string]interface{}{
				"group_id":            form.GroupID,
				"group_slug":          shieldGroup.Slug,
				"resource_type":       form.ResourceType,
				"resource_id":         form.ResourceID,
				"project_id":          form.ProjectID,
				"project_slug":        shieldProject.Slug,
				"channel_criticality": string(form.ChannelCriticality),
				"channel_name":        channelName,
			}),
			CreatedBy: form.UserID,
		}

		shield := new(mocks.ShieldServiceClient)
		shield.On("GetProject", ctx, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		shield.On("GetGroup", ctx, &shieldv1beta1.GetGroupRequest{Id: form.GroupID}).
			Return(&shieldv1beta1.GetGroupResponse{
				Group: shieldGroup,
			}, nil)
		defer shield.AssertExpectations(t)
		siren := new(mocks.SirenServiceClient)
		siren.On("ListReceivers", ctx, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team":     shieldGroup.Slug,
				"severity": string(form.ChannelCriticality),
			},
		}).Return(&sirenv1beta1.ListReceiversResponse{Receivers: sirenReceivers}, nil)
		siren.
			On("CreateSubscription", ctx, expectedSirenPayload).
			Return(&sirenv1beta1.CreateSubscriptionResponse{Id: 5}, nil)
		defer siren.AssertExpectations(t)

		service := alert.NewSubscriptionService(siren, shield)
		subsID, err := service.CreateSubscription(ctx, form)
		assert.NoError(t, err)
		assert.Equal(t, 5, subsID)
	})
}

func TestSubscriptionServiceUpdateSubscription(t *testing.T) {
	var (
		ctx            = context.TODO()
		subscriptionID = 205
		groupID        = "8a7219cd-53c9-47f1-9387-5cac7abe4dcb"
		projectID      = "5dab4194-9516-421a-aafe-72fd3d96ec56"
	)

	t.Run("should return error if siren namespace cannot be retrieved from project", func(t *testing.T) {
		tests := []struct {
			name     string
			metadata *structpb.Struct
		}{
			{
				name:     "empty metadata",
				metadata: nil,
			},
			{
				name: "empty metadata.siren_namespace",
				metadata: newStruct(t, map[string]interface{}{
					"siren_namespace": nil,
				}),
			},
			{
				name: "invalid format for metadata.siren_namespace",
				metadata: newStruct(t, map[string]interface{}{
					"siren_namespace": "wrong-format",
				}),
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				form := alert.SubscriptionForm{
					ProjectID: projectID,
					GroupID:   groupID,
				}
				shieldProject := &shieldv1beta1.Project{
					Slug:     "test-project",
					Metadata: test.metadata,
				}

				shield := new(mocks.ShieldServiceClient)
				shield.On("GetProject", ctx, &shieldv1beta1.GetProjectRequest{Id: projectID}).
					Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
				defer shield.AssertExpectations(t)
				client := new(mocks.SirenServiceClient)
				defer client.AssertExpectations(t)

				service := alert.NewSubscriptionService(client, shield)
				err := service.UpdateSubscription(ctx, subscriptionID, form)
				assert.ErrorIs(t, err, alert.ErrNoShieldSirenNamespace)
			})
		}
	})

	t.Run("should return error on failing to get siren's receiver", func(t *testing.T) {
		tests := []struct {
			name          string
			receivers     []*sirenv1beta1.Receiver
			expectedError error
		}{
			{
				name:          "nil receivers",
				receivers:     nil,
				expectedError: alert.ErrNoSirenReceiver,
			},
			{
				name:          "empty receivers",
				receivers:     []*sirenv1beta1.Receiver{},
				expectedError: alert.ErrNoSirenReceiver,
			},
			{
				name: "more than one receivers",
				receivers: []*sirenv1beta1.Receiver{
					{Id: 1},
					{Id: 2},
				},
			},
			{
				name: "receiver is not slack_channel type",
				receivers: []*sirenv1beta1.Receiver{
					{Id: 1, Type: "invalid-type"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				form := alert.SubscriptionForm{
					ProjectID:          projectID,
					GroupID:            groupID,
					ChannelCriticality: alert.ChannelCriticalityInfo,
				}
				shieldGroup := &shieldv1beta1.Group{
					Slug: "test-group",
				}
				shieldProject := &shieldv1beta1.Project{
					Slug: "test-project",
					Metadata: newStruct(t, map[string]interface{}{
						"siren_namespace": 5,
					}),
				}

				shield := new(mocks.ShieldServiceClient)
				shield.On("GetProject", ctx, &shieldv1beta1.GetProjectRequest{Id: projectID}).
					Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
				shield.On("GetGroup", ctx, &shieldv1beta1.GetGroupRequest{Id: form.GroupID}).
					Return(&shieldv1beta1.GetGroupResponse{Group: shieldGroup}, nil)
				defer shield.AssertExpectations(t)
				siren := new(mocks.SirenServiceClient)
				siren.On("ListReceivers", ctx, &sirenv1beta1.ListReceiversRequest{
					Labels: map[string]string{
						"team":     shieldGroup.Slug,
						"severity": string(form.ChannelCriticality),
					},
				}).Return(&sirenv1beta1.ListReceiversResponse{
					Receivers: test.receivers,
				}, nil)
				defer siren.AssertExpectations(t)

				service := alert.NewSubscriptionService(siren, shield)
				err := service.UpdateSubscription(ctx, subscriptionID, form)
				if test.expectedError != nil {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("should update subscription on success", func(t *testing.T) {
		receiverID := uint64(17)
		sirenNamespace := 5
		channelName := "test-channel-update"

		// inputs
		form := alert.SubscriptionForm{
			UserID:             "john.doe@example.com",
			AlertSeverity:      alert.AlertSeverityCritical,
			ChannelCriticality: alert.ChannelCriticalityInfo,
			GroupID:            groupID,
			ProjectID:          projectID,
			ResourceType:       "firehose",
			ResourceID:         "test-job",
		}

		// conditions
		shieldGroup := &shieldv1beta1.Group{
			Slug: "test-group",
		}
		shieldProject := &shieldv1beta1.Project{
			Slug: "my-project-1",
			Metadata: newStruct(t, map[string]interface{}{
				"siren_namespace": sirenNamespace,
			}),
		}
		sirenReceivers := []*sirenv1beta1.Receiver{
			{Id: receiverID, Type: sirenReceiverPkg.TypeSlackChannel, Configurations: newStruct(t, map[string]interface{}{
				"channel_name": channelName,
			})},
		}

		// expecations
		expectedSirenPayload := &sirenv1beta1.UpdateSubscriptionRequest{
			Id: uint64(subscriptionID),
			Urn: fmt.Sprintf(
				"%s:%s:%s:%s",
				shieldGroup.GetSlug(), form.AlertSeverity, form.ResourceType, form.ResourceID,
			),
			Namespace: uint64(sirenNamespace),
			Receivers: []*sirenv1beta1.ReceiverMetadata{
				{Id: receiverID},
			},
			Match: map[string]string{
				"severity":   string(alert.AlertSeverityCritical),
				"identifier": "test-job",
			},
			Metadata: newStruct(t, map[string]interface{}{
				"group_id":            form.GroupID,
				"group_slug":          shieldGroup.Slug,
				"resource_type":       form.ResourceType,
				"resource_id":         form.ResourceID,
				"project_id":          form.ProjectID,
				"project_slug":        shieldProject.Slug,
				"channel_criticality": string(form.ChannelCriticality),
				"channel_name":        channelName,
			}),
			UpdatedBy: form.UserID,
		}

		shield := new(mocks.ShieldServiceClient)
		shield.On("GetProject", ctx, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		shield.On("GetGroup", ctx, &shieldv1beta1.GetGroupRequest{Id: form.GroupID}).
			Return(&shieldv1beta1.GetGroupResponse{
				Group: shieldGroup,
			}, nil)
		defer shield.AssertExpectations(t)
		siren := new(mocks.SirenServiceClient)
		siren.On("ListReceivers", ctx, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team":     shieldGroup.Slug,
				"severity": string(form.ChannelCriticality),
			},
		}).Return(&sirenv1beta1.ListReceiversResponse{Receivers: sirenReceivers}, nil)
		siren.
			On("UpdateSubscription", ctx, expectedSirenPayload).
			Return(&sirenv1beta1.UpdateSubscriptionResponse{}, nil)
		defer siren.AssertExpectations(t)

		service := alert.NewSubscriptionService(siren, shield)
		err := service.UpdateSubscription(ctx, subscriptionID, form)
		assert.NoError(t, err)
	})
}

func TestSubscriptionServiceDeleteSubscription(t *testing.T) {
	ctx := context.TODO()
	subscriptionID := 203

	t.Run("should not return error success", func(t *testing.T) {
		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("DeleteSubscription", ctx, &sirenv1beta1.DeleteSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, nil)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		err := service.DeleteSubscription(ctx, subscriptionID)
		assert.NoError(t, err)
	})

	t.Run("should return not found error if optimus return NotFound code", func(t *testing.T) {
		expectedError := status.Error(codes.NotFound, "Not Found")

		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("DeleteSubscription", ctx, &sirenv1beta1.DeleteSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, expectedError)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		err := service.DeleteSubscription(ctx, subscriptionID)
		assert.ErrorIs(t, err, alert.ErrSubscriptionNotFound)
	})

	t.Run("should return if client return error", func(t *testing.T) {
		expectedError := status.Error(codes.Internal, "Internal")

		shield := new(mocks.ShieldServiceClient)
		client := new(mocks.SirenServiceClient)
		client.On("DeleteSubscription", ctx, &sirenv1beta1.DeleteSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, expectedError)
		defer client.AssertExpectations(t)

		service := alert.NewSubscriptionService(client, shield)
		err := service.DeleteSubscription(ctx, subscriptionID)
		assert.ErrorIs(t, err, expectedError)
	})
}

func TestSubscriptionServiceGetAlertChannels(t *testing.T) {
	ctx := context.TODO()
	groupID := "deafcced-845c-4089-89f0-06621486cb0a"

	t.Run("should not return error if group is not found", func(t *testing.T) {
		notFoundError := status.Error(codes.NotFound, "Not Found")

		shield := new(mocks.ShieldServiceClient)
		shield.On("GetGroup", ctx, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(nil, notFoundError)
		defer shield.AssertExpectations(t)
		siren := new(mocks.SirenServiceClient)

		service := alert.NewSubscriptionService(siren, shield)
		_, err := service.GetAlertChannels(ctx, groupID)
		assert.ErrorIs(t, err, alert.ErrNoShieldGroup)
	})

	t.Run("should return alert channels", func(t *testing.T) {
		groupSlug := "test-group-30"
		shieldGroup := &shieldv1beta1.Group{
			Slug: groupSlug,
		}
		sirenReceivers := []*sirenv1beta1.Receiver{
			{
				Id:   54,
				Name: "test-receiver-info-1",
				Labels: map[string]string{
					"severity": string(alert.AlertSeverityInfo),
				},
				Configurations: newStruct(t, map[string]interface{}{
					"channel_name": "test-channel-info-1",
				}),
			},
			{
				Id:   55,
				Name: "test-receiver-critical-1",
				Labels: map[string]string{
					"severity": string(alert.AlertSeverityCritical),
				},
				Configurations: newStruct(t, map[string]interface{}{
					"channel_name": "test-channel-critical-1",
				}),
			},
			{
				Id:   56,
				Name: "test-receiver-warning-1",
				Labels: map[string]string{
					"severity": string(alert.AlertSeverityWarning),
				},
				Configurations: newStruct(t, map[string]interface{}{
					"channel_name": "test-channel-warning-1",
				}),
			},
		}

		expected := []models.AlertChannel{
			{
				ReceiverID:         fmt.Sprint(sirenReceivers[0].Id),
				ReceiverName:       sirenReceivers[0].Name,
				ChannelCriticality: models.NewChannelCriticality(models.ChannelCriticalityINFO),
				ChannelName:        "test-channel-info-1",
			},
			{
				ReceiverID:         fmt.Sprint(sirenReceivers[1].Id),
				ReceiverName:       sirenReceivers[1].Name,
				ChannelCriticality: models.NewChannelCriticality(models.ChannelCriticalityCRITICAL),
				ChannelName:        "test-channel-critical-1",
			},
			{
				ReceiverID:         fmt.Sprint(sirenReceivers[2].Id),
				ReceiverName:       sirenReceivers[2].Name,
				ChannelCriticality: models.NewChannelCriticality(models.ChannelCriticalityWARNING),
				ChannelName:        "test-channel-warning-1",
			},
		}

		shield := new(mocks.ShieldServiceClient)
		shield.On("GetGroup", ctx, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(&shieldv1beta1.GetGroupResponse{
				Group: shieldGroup,
			}, nil)
		defer shield.AssertExpectations(t)
		siren := new(mocks.SirenServiceClient)
		siren.On("ListReceivers", ctx, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team": groupSlug,
			},
		}).
			Return(&sirenv1beta1.ListReceiversResponse{
				Receivers: sirenReceivers,
			}, nil)
		defer siren.AssertExpectations(t)

		service := alert.NewSubscriptionService(siren, shield)
		actual, err := service.GetAlertChannels(ctx, groupID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func newStruct(t *testing.T, d map[string]interface{}) *structpb.Struct {
	t.Helper()

	strct, err := structpb.NewStruct(d)
	require.NoError(t, err)
	return strct
}
