package optimus_test

import (
	"context"
	"testing"

	optimusv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/optimus/core/v1beta1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/internal/server/v1/optimus"
	"github.com/goto/dex/mocks"
	"github.com/goto/dex/pkg/errors"
)

func TestServiceFindJobSpec(t *testing.T) {
	jobName := "sample-optimus-job-name"
	projectName := "sample-project"

	t.Run("should return job spec using job name and project name from argument", func(t *testing.T) {
		jobSpecRes := &optimusv1beta1.JobSpecificationResponse{
			ProjectName:   "test-project",
			NamespaceName: "test-namespcace",
			Job: &optimusv1beta1.JobSpecification{
				Version:  1,
				Name:     "sample-job",
				Owner:    "goto",
				TaskName: "sample-task-name",
			},
		}

		client := new(mocks.JobSpecificationServiceClient)
		client.On("GetJobSpecifications", context.TODO(), &optimusv1beta1.GetJobSpecificationsRequest{
			ProjectName: projectName,
			JobName:     jobName,
		}).Return(&optimusv1beta1.GetJobSpecificationsResponse{
			JobSpecificationResponses: []*optimusv1beta1.JobSpecificationResponse{
				jobSpecRes,
			},
		}, nil)
		defer client.AssertExpectations(t)

		service := optimus.NewService(client)
		job, err := service.FindJobSpec(context.TODO(), jobName, projectName)
		assert.NoError(t, err)
		assert.Equal(t, jobSpecRes, job)
	})

	t.Run("should return not found, if job could not be found", func(t *testing.T) {
		client := new(mocks.JobSpecificationServiceClient)
		client.On("GetJobSpecifications", context.TODO(), &optimusv1beta1.GetJobSpecificationsRequest{
			ProjectName: projectName,
			JobName:     jobName,
		}).Return(&optimusv1beta1.GetJobSpecificationsResponse{}, nil)
		defer client.AssertExpectations(t)

		service := optimus.NewService(client)
		_, err := service.FindJobSpec(context.TODO(), jobName, projectName)
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("should return error, if client fails", func(t *testing.T) {
		expectedErr := status.Error(codes.Internal, "Internal")

		client := new(mocks.JobSpecificationServiceClient)
		client.On("GetJobSpecifications", context.TODO(), &optimusv1beta1.GetJobSpecificationsRequest{
			ProjectName: projectName,
			JobName:     jobName,
		}).Return(nil, expectedErr)
		defer client.AssertExpectations(t)

		service := optimus.NewService(client)
		_, err := service.FindJobSpec(context.TODO(), jobName, projectName)
		assert.ErrorIs(t, err, expectedErr)
	})
}
