package alert_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"
	"github.com/go-chi/chi/v5"
	sirenReceiverPkg "github.com/goto/siren/core/receiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/reqctx"
	"github.com/goto/dex/internal/server/v1/alert"
	"github.com/goto/dex/mocks"
	"github.com/goto/dex/pkg/errors"
)

const (
	emailHeaderKey = "X-Auth-Email"
)

func TestRoutesFindSubscription(t *testing.T) {
	subscriptionID := 102
	path := "/102"
	method := http.MethodGet

	t.Run("should return subscription on success", func(t *testing.T) {
		subscription := &sirenv1beta1.Subscription{
			Id:        uint64(subscriptionID),
			Urn:       "sample-http-call-urn",
			Namespace: 1,
			Receivers: []*sirenv1beta1.ReceiverMetadata{
				{Id: 32},
			},
		}

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("GetSubscription", mock.Anything, &sirenv1beta1.GetSubscriptionRequest{Id: subscription.Id}).
			Return(&sirenv1beta1.GetSubscriptionResponse{
				Subscription: subscription,
			}, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusOK, response.Code)

		// assert response
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(map[string]interface{}{
			"subscription": alert.MapToSubscription(subscription),
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})

	t.Run("should return 404 if id is not found", func(t *testing.T) {
		expectedError := status.Error(codes.NotFound, "not found")

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("GetSubscription", mock.Anything, &sirenv1beta1.GetSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, expectedError)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		expectedStatusCode := http.StatusNotFound
		assert.Equal(t, expectedStatusCode, response.Code)
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(errors.Error{
			Status:  expectedStatusCode,
			Message: alert.ErrSubscriptionNotFound.Error(),
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})

	t.Run("should return 500 for internal error", func(t *testing.T) {
		expectedError := status.Error(codes.Internal, "Internal")

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("GetSubscription", mock.Anything, &sirenv1beta1.GetSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, expectedError)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		expectedStatusCode := http.StatusInternalServerError
		assert.Equal(t, expectedStatusCode, response.Code)
	})
}

func TestRoutesGetSubscriptions(t *testing.T) {
	groupID := "sample-shield-group-id"
	resourceID := "sample-dagger-id-or-urn"
	resourceType := "dagger"
	method := http.MethodGet

	t.Run("should return subscription on success", func(t *testing.T) {
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

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("ListSubscriptions", mock.Anything, &sirenv1beta1.ListSubscriptionsRequest{Metadata: map[string]string{
			"group_id":      groupID,
			"resource_id":   resourceID,
			"resource_type": resourceType,
		}}).
			Return(&sirenv1beta1.ListSubscriptionsResponse{
				Subscriptions: subscriptions,
			}, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(
			method,
			fmt.Sprintf("/?group_id=%s&resource_id=%s&resource_type=%s", groupID, resourceID, resourceType),
			nil,
		)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusOK, response.Code)

		// assert response
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(map[string]interface{}{
			"subscriptions": alert.MapToSubscriptionList(subscriptions),
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})

	t.Run("should return 400 if both group_id and resource_id is not passed", func(t *testing.T) {
		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/", nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("should return 500 on internal error", func(t *testing.T) {
		expectedError := status.Error(codes.Internal, "Internal")

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("ListSubscriptions", mock.Anything, &sirenv1beta1.ListSubscriptionsRequest{Metadata: map[string]string{
			"group_id":      groupID,
			"resource_id":   resourceID,
			"resource_type": resourceType,
		}}).
			Return(nil, expectedError)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(
			method,
			fmt.Sprintf("/?group_id=%s&resource_id=%s&resource_type=%s", groupID, resourceID, resourceType),
			nil,
		)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestRoutesCreateSubscriptions(t *testing.T) {
	var (
		method             = http.MethodPost
		channelCriticality = alert.ChannelCriticalityInfo
		projectID          = "5dab4194-9516-421a-aafe-72fd3d96ec56"
		groupID            = "8a7219cd-53c9-47f1-9387-5cac7abe4dcb"
		userEmail          = "jane.doe@example.com"
		validJSONPayload   = fmt.Sprintf(`{
			"project_id": "%s",
      "resource_id": "test-pipeline-job",
      "resource_type": "optimus",
      "group_id": "%s",
      "alert_severity": "CRITICAL",
      "channel_criticality": "%s"
		}`, projectID, groupID, channelCriticality)
	)

	t.Run("should return 401 on empty user header", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/", requestBody)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	})

	t.Run("should return 400 on validation error", func(t *testing.T) {
		tests := []struct {
			jsonString string
		}{
			{jsonString: ``},
			{jsonString: `{}`},
			{jsonString: `{
				"project_id": "",
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "CRITICAL",
				"channel_criticality": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "critical",
				"channel_criticality": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "CRITICAL",
				"channel_criticality": "info"
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "critical",
				"channel_criticality": "INFO"
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "critical",
				"channel_criticality": "info"
			}`},
		}

		for i, test := range tests {
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				requestBody := strings.NewReader(test.jsonString)

				shieldClient := new(mocks.ShieldServiceClient)
				sirenClient := new(mocks.SirenServiceClient)

				response := httptest.NewRecorder()
				request := httptest.NewRequest(method, "/", requestBody)
				request.Header.Set(emailHeaderKey, userEmail)
				router := getRouter()
				alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
				router.ServeHTTP(response, request)

				// assert status
				assert.Equal(t, http.StatusBadRequest, response.Code)
			})
		}
	})

	t.Run("should return 422 on namespace could not be found on shield project", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)

		shieldProject := &shieldv1beta1.Project{
			Slug:     "test-project",
			Metadata: nil,
		}
		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetProject", mock.Anything, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/", requestBody)
		request.Header.Set(emailHeaderKey, userEmail)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	})

	t.Run("should return 422 on receiver could not be found", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)

		shieldProject := &shieldv1beta1.Project{
			Slug: "test-project",
			Metadata: newStruct(t, map[string]interface{}{
				"siren_namespace": 3,
			}),
		}
		shieldGroup := &shieldv1beta1.Group{
			Slug: "test-group",
		}
		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetProject", mock.Anything, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		shieldClient.On("GetGroup", mock.Anything, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(&shieldv1beta1.GetGroupResponse{Group: shieldGroup}, nil)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("ListReceivers", mock.Anything, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team":     shieldGroup.Slug,
				"severity": string(channelCriticality),
			},
		}).Return(&sirenv1beta1.ListReceiversResponse{Receivers: nil}, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/", requestBody)
		request.Header.Set(emailHeaderKey, userEmail)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	})

	t.Run("should return 201 on success", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)
		sirenNamespace := 13
		receiverID := uint64(25)
		subscriptionID := 200
		channelName := "test-channel-30"

		shieldProject := &shieldv1beta1.Project{
			Slug: "test-project",
			Metadata: newStruct(t, map[string]interface{}{
				"siren_namespace": sirenNamespace,
			}),
		}
		shieldGroup := &shieldv1beta1.Group{
			Slug: "test-group",
		}
		sirenReceivers := []*sirenv1beta1.Receiver{
			{Id: receiverID, Type: sirenReceiverPkg.TypeSlackChannel, Configurations: newStruct(t, map[string]interface{}{
				"channel_name": channelName,
			})},
		}

		expectedSirenPayload := &sirenv1beta1.CreateSubscriptionRequest{
			Urn: fmt.Sprintf(
				"%s:%s:%s:%s",
				"test-group", "CRITICAL", "optimus", "test-pipeline-job",
			),
			Namespace: uint64(sirenNamespace),
			Receivers: []*sirenv1beta1.ReceiverMetadata{
				{Id: receiverID},
			},
			Match: map[string]string{
				"severity":   "CRITICAL",
				"identifier": "test-pipeline-job",
			},
			Metadata: newStruct(t, map[string]interface{}{
				"group_id":            groupID,
				"group_slug":          shieldGroup.Slug,
				"resource_type":       "optimus",
				"resource_id":         "test-pipeline-job",
				"project_id":          projectID,
				"project_slug":        shieldProject.Slug,
				"channel_criticality": string(channelCriticality),
				"channel_name":        channelName,
			}),
			CreatedBy: userEmail,
		}
		sirenSubscription := &sirenv1beta1.Subscription{
			Namespace: uint64(sirenNamespace),
		}

		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetProject", mock.Anything, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		shieldClient.On("GetGroup", mock.Anything, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(&shieldv1beta1.GetGroupResponse{Group: shieldGroup}, nil)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("ListReceivers", mock.Anything, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team":     shieldGroup.Slug,
				"severity": string(channelCriticality),
			},
		}).Return(&sirenv1beta1.ListReceiversResponse{Receivers: sirenReceivers}, nil)
		sirenClient.
			On("CreateSubscription", mock.Anything, expectedSirenPayload).
			Return(&sirenv1beta1.CreateSubscriptionResponse{Id: uint64(subscriptionID)}, nil)
		sirenClient.
			On("GetSubscription", mock.Anything, &sirenv1beta1.GetSubscriptionRequest{
				Id: uint64(subscriptionID),
			}).
			Return(&sirenv1beta1.GetSubscriptionResponse{
				Subscription: sirenSubscription,
			}, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/", requestBody)
		request.Header.Set(emailHeaderKey, userEmail)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		assert.Equal(t, http.StatusCreated, response.Code)
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(map[string]interface{}{
			"subscription": alert.MapToSubscription(sirenSubscription),
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})
}

func TestRoutesUpdateSubscriptions(t *testing.T) {
	var (
		method             = http.MethodPut
		subscriptionID     = 305
		urlPath            = fmt.Sprintf("/%d", subscriptionID)
		channelCriticality = alert.ChannelCriticalityInfo
		projectID          = "5dab4194-9516-421a-aafe-72fd3d96ec56"
		groupID            = "8a7219cd-53c9-47f1-9387-5cac7abe4dcb"
		userEmail          = "jane.doe@example.com"
		validJSONPayload   = fmt.Sprintf(`{
			"project_id": "%s",
      "resource_id": "test-pipeline-job",
      "resource_type": "optimus",
      "group_id": "%s",
      "alert_severity": "CRITICAL",
      "channel_criticality": "%s"
		}`, projectID, groupID, channelCriticality)
	)

	t.Run("should return 401 on empty user header", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, urlPath, requestBody)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	})

	t.Run("should return 400 on validation error", func(t *testing.T) {
		tests := []struct {
			jsonString string
		}{
			{jsonString: ``},
			{jsonString: `{}`},
			{jsonString: `{
				"project_id": "",
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "CRITICAL",
				"channel_criticality": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "critical",
				"channel_criticality": ""
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "CRITICAL",
				"channel_criticality": "info"
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "critical",
				"channel_criticality": "INFO"
			}`},
			{jsonString: `{
				"project_id": "5dab4194-9516-421a-aafe-72fd3d96ec56",
				"resource_id": "test-pipeline-job",
				"resource_type": "optimus",
				"group_id": "8a7219cd-53c9-47f1-9387-5cac7abe4dcb",
				"alert_severity": "critical",
				"channel_criticality": "info"
			}`},
		}

		for i, test := range tests {
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				requestBody := strings.NewReader(test.jsonString)

				shieldClient := new(mocks.ShieldServiceClient)
				sirenClient := new(mocks.SirenServiceClient)

				response := httptest.NewRecorder()
				request := httptest.NewRequest(method, urlPath, requestBody)
				request.Header.Set(emailHeaderKey, userEmail)
				router := getRouter()
				alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
				router.ServeHTTP(response, request)

				// assert status
				assert.Equal(t, http.StatusBadRequest, response.Code)
			})
		}
	})

	t.Run("should return 422 on namespace could not be found on shield project", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)

		shieldProject := &shieldv1beta1.Project{
			Slug:     "test-project",
			Metadata: nil,
		}
		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetProject", mock.Anything, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, urlPath, requestBody)
		request.Header.Set(emailHeaderKey, userEmail)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	})

	t.Run("should return 422 on receiver could not be found", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)

		shieldProject := &shieldv1beta1.Project{
			Slug: "test-project",
			Metadata: newStruct(t, map[string]interface{}{
				"siren_namespace": 3,
			}),
		}
		shieldGroup := &shieldv1beta1.Group{
			Slug: "test-group",
		}
		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetProject", mock.Anything, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		shieldClient.On("GetGroup", mock.Anything, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(&shieldv1beta1.GetGroupResponse{Group: shieldGroup}, nil)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("ListReceivers", mock.Anything, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team":     shieldGroup.Slug,
				"severity": string(channelCriticality),
			},
		}).Return(&sirenv1beta1.ListReceiversResponse{Receivers: nil}, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, urlPath, requestBody)
		request.Header.Set(emailHeaderKey, userEmail)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	})

	t.Run("should return 200 on success", func(t *testing.T) {
		requestBody := strings.NewReader(validJSONPayload)
		receiverID := uint64(30)
		sirenNamespace := 13
		channelName := "test-channel-70"

		shieldProject := &shieldv1beta1.Project{
			Slug: "test-project",
			Metadata: newStruct(t, map[string]interface{}{
				"siren_namespace": sirenNamespace,
			}),
		}
		shieldGroup := &shieldv1beta1.Group{
			Slug: "test-group",
		}
		sirenReceivers := []*sirenv1beta1.Receiver{
			{Id: receiverID, Type: sirenReceiverPkg.TypeSlackChannel, Configurations: newStruct(t, map[string]interface{}{
				"channel_name": channelName,
			})},
		}

		expectedSirenPayload := &sirenv1beta1.UpdateSubscriptionRequest{
			Id: uint64(subscriptionID),
			Urn: fmt.Sprintf(
				"%s:%s:%s:%s",
				shieldGroup.Slug, "CRITICAL", "optimus", "test-pipeline-job",
			),
			Namespace: uint64(sirenNamespace),
			Receivers: []*sirenv1beta1.ReceiverMetadata{
				{Id: receiverID},
			},
			Match: map[string]string{
				"severity":   "CRITICAL",
				"identifier": "test-pipeline-job",
			},
			Metadata: newStruct(t, map[string]interface{}{
				"group_id":            groupID,
				"group_slug":          shieldGroup.Slug,
				"resource_type":       "optimus",
				"resource_id":         "test-pipeline-job",
				"project_id":          projectID,
				"project_slug":        shieldProject.Slug,
				"channel_criticality": string(channelCriticality),
				"channel_name":        channelName,
			}),
			UpdatedBy: userEmail,
		}
		sirenSubscription := &sirenv1beta1.Subscription{
			Namespace: uint64(sirenNamespace),
		}

		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetProject", mock.Anything, &shieldv1beta1.GetProjectRequest{Id: projectID}).
			Return(&shieldv1beta1.GetProjectResponse{Project: shieldProject}, nil)
		shieldClient.On("GetGroup", mock.Anything, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(&shieldv1beta1.GetGroupResponse{Group: shieldGroup}, nil)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("ListReceivers", mock.Anything, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team":     shieldGroup.Slug,
				"severity": string(channelCriticality),
			},
		}).Return(&sirenv1beta1.ListReceiversResponse{Receivers: sirenReceivers}, nil)
		sirenClient.
			On("UpdateSubscription", mock.Anything, expectedSirenPayload).
			Return(nil, nil)
		sirenClient.
			On("GetSubscription", mock.Anything, &sirenv1beta1.GetSubscriptionRequest{
				Id: uint64(subscriptionID),
			}).
			Return(&sirenv1beta1.GetSubscriptionResponse{
				Subscription: sirenSubscription,
			}, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, urlPath, requestBody)
		request.Header.Set(emailHeaderKey, userEmail)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(map[string]interface{}{
			"subscription": alert.MapToSubscription(sirenSubscription),
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})
}

func TestRoutesDeleteSubscription(t *testing.T) {
	subscriptionID := 202
	path := "/202"
	method := http.MethodDelete

	t.Run("should return 200 on success", func(t *testing.T) {
		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("DeleteSubscription", mock.Anything, &sirenv1beta1.DeleteSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("should return 404 if id is not found", func(t *testing.T) {
		expectedError := status.Error(codes.NotFound, "not found")

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("DeleteSubscription", mock.Anything, &sirenv1beta1.DeleteSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, expectedError)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		expectedStatusCode := http.StatusNotFound
		assert.Equal(t, expectedStatusCode, response.Code)
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(errors.Error{
			Status:  expectedStatusCode,
			Message: alert.ErrSubscriptionNotFound.Error(),
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})

	t.Run("should return 500 for internal error", func(t *testing.T) {
		expectedError := status.Error(codes.Internal, "Internal")

		shieldClient := new(mocks.ShieldServiceClient)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("DeleteSubscription", mock.Anything, &sirenv1beta1.DeleteSubscriptionRequest{Id: uint64(subscriptionID)}).
			Return(nil, expectedError)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		expectedStatusCode := http.StatusInternalServerError
		assert.Equal(t, expectedStatusCode, response.Code)
	})
}

func TestRoutesGetAlertChannels(t *testing.T) {
	var (
		method  = http.MethodGet
		groupID = "8a7219cd-53c9-47f1-9387-5cac7abe4dcb"
		urlPath = fmt.Sprintf("/groups/%s/alert_channels", groupID)
	)

	t.Run("should return 404 on group not found", func(t *testing.T) {
		notFoundError := status.Error(codes.NotFound, "Not Found")

		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetGroup", mock.Anything, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(nil, notFoundError)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, urlPath, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert status
		assert.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("should return 200 and alert channels on success", func(t *testing.T) {
		groupSlug := "test-project"
		channelName := "test-channel-info-2"

		shieldGroup := &shieldv1beta1.Group{
			Slug: groupSlug,
		}
		sirenReceivers := []*sirenv1beta1.Receiver{
			{
				Id:   30,
				Name: "test-receiver-info-2",
				Labels: map[string]string{
					"severity": string(alert.AlertSeverityInfo),
				},
				Configurations: newStruct(t, map[string]interface{}{
					"channel_name": channelName,
				}),
			},
		}
		alertChannels := []models.AlertChannel{
			{
				ReceiverID:         fmt.Sprint(sirenReceivers[0].Id),
				ReceiverName:       sirenReceivers[0].Name,
				ChannelCriticality: models.NewChannelCriticality(models.ChannelCriticalityINFO),
				ChannelName:        channelName,
			},
		}

		shieldClient := new(mocks.ShieldServiceClient)
		shieldClient.On("GetGroup", mock.Anything, &shieldv1beta1.GetGroupRequest{Id: groupID}).
			Return(&shieldv1beta1.GetGroupResponse{
				Group: shieldGroup,
			}, nil)
		defer shieldClient.AssertExpectations(t)
		sirenClient := new(mocks.SirenServiceClient)
		sirenClient.On("ListReceivers", mock.Anything, &sirenv1beta1.ListReceiversRequest{
			Labels: map[string]string{
				"team": groupSlug,
			},
		}).
			Return(&sirenv1beta1.ListReceiversResponse{
				Receivers: sirenReceivers,
			}, nil)
		defer sirenClient.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, urlPath, nil)
		router := getRouter()
		alert.SubscriptionRoutes(sirenClient, shieldClient)(router)
		router.ServeHTTP(response, request)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(map[string]interface{}{
			"alert_channels": alertChannels,
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})
}

func getRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(reqctx.WithRequestCtx())

	return router
}
