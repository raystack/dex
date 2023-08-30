package dlq

import (
	entropyv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
)

type Service struct {
	client entropyv1beta1rpc.ResourceServiceClient
}

func NewService(client entropyv1beta1rpc.ResourceServiceClient) *Service {
	return &Service{
		client: client,
	}
}
