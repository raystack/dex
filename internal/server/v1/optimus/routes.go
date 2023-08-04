package optimus

import (
	optimusv1beta1grpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/optimus/core/v1beta1/corev1beta1grpc"
	"github.com/go-chi/chi/v5"
)

func Routes(optimusClient optimusv1beta1grpc.JobSpecificationServiceClient) func(r chi.Router) {
	service := NewService(optimusClient)
	handler := NewHandler(service)

	return func(r chi.Router) {
		r.Get("/projects/{project_name}/jobs/{job_name}", handler.findJob)
		r.Get("/projects/{project_name}/optimus", handler.list)
	}

}
