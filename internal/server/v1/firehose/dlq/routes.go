package dlq

import (
	entropyv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
	"github.com/go-chi/chi/v5"
)

func Routes(entropyClient entropyv1beta1rpc.ResourceServiceClient) func(r chi.Router) {
	service := NewService(entropyClient)
	handler := NewHandler(service)

	return func(r chi.Router) {
		r.Get("/", handler.listDlq)
		r.Get("/jobs", handler.listDlqJobs)
		r.Get("/jobs/{job_urn}", handler.getDlqJob)
		r.Post("/jobs", handler.createDlqJob)
	}
}
