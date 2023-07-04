package optimus_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	optimusv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/optimus/core/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/internal/server/v1/optimus"
	"github.com/goto/dex/mocks"
)

func TestRoutesFindJobSpec(t *testing.T) {
	jobName := "sample-optimus-job-name"
	projectName := "sample-project"
	method := http.MethodGet
	path := fmt.Sprintf("/projects/%s/jobs/%s", projectName, jobName)

	t.Run("should return 200 with job spec", func(t *testing.T) {
		jobSpec := &optimusv1beta1.JobSpecification{
			Version:  1,
			Name:     "sample-job",
			Owner:    "goto",
			TaskName: "sample-task-name",
		}

		client := new(mocks.JobSpecificationServiceClient)
		client.On("GetJobSpecifications", mock.Anything, &optimusv1beta1.GetJobSpecificationsRequest{
			ProjectName: projectName,
			JobName:     jobName,
		}).Return(&optimusv1beta1.GetJobSpecificationsResponse{
			JobSpecificationResponses: []*optimusv1beta1.JobSpecificationResponse{
				{Job: jobSpec},
			},
		}, nil)
		defer client.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := chi.NewRouter()
		optimus.Routes(client)(router)
		router.ServeHTTP(response, request)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		resultJSON := response.Body.Bytes()
		expectedJSON, err := json.Marshal(map[string]interface{}{
			"job": jobSpec,
		})
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(resultJSON))
	})

	t.Run("should return 404 if job could not be found", func(t *testing.T) {
		client := new(mocks.JobSpecificationServiceClient)
		client.On("GetJobSpecifications", mock.Anything, &optimusv1beta1.GetJobSpecificationsRequest{
			ProjectName: projectName,
			JobName:     jobName,
		}).Return(&optimusv1beta1.GetJobSpecificationsResponse{}, nil)
		defer client.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := chi.NewRouter()
		optimus.Routes(client)(router)
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("should return 500 for internal error", func(t *testing.T) {
		clientError := status.Error(codes.Internal, "Internal")

		client := new(mocks.JobSpecificationServiceClient)
		client.On("GetJobSpecifications", mock.Anything, &optimusv1beta1.GetJobSpecificationsRequest{
			ProjectName: projectName,
			JobName:     jobName,
		}).Return(nil, clientError)
		defer client.AssertExpectations(t)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, nil)
		router := chi.NewRouter()
		optimus.Routes(client)(router)
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}
