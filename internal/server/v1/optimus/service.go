package optimus

import (
	"context"

	optimusv1beta1grpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/optimus/core/v1beta1/corev1beta1grpc"
	optimusv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/optimus/core/v1beta1"

	"github.com/goto/dex/pkg/errors"
)

type Service struct {
	client optimusv1beta1grpc.JobSpecificationServiceClient
}

func NewService(client optimusv1beta1grpc.JobSpecificationServiceClient) *Service {
	return &Service{
		client: client,
	}
}

func (svc *Service) FindJobSpec(ctx context.Context, jobName, projectName string) (*optimusv1beta1.JobSpecificationResponse, error) {
	res, err := svc.client.GetJobSpecifications(ctx, &optimusv1beta1.GetJobSpecificationsRequest{
		ProjectName: projectName,
		JobName:     jobName,
	})
	if err != nil {
		return nil, err
	}

	list := res.JobSpecificationResponses
	if len(list) == 0 {
		return nil, errors.ErrNotFound
	}

	return list[0], nil
}

func (svc *Service) ListJobs(ctx context.Context, projectName string) (*optimusv1beta1.ListJobSpecificationResponse, error) {
	res, err := svc.client.ListJobSpecification(ctx, &optimusv1beta1.ListJobSpecificationRequest{
		ProjectName:   projectName,
		NamespaceName: "smoke_test",
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
