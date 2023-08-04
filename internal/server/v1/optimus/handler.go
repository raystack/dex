package optimus

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goto/dex/internal/server/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) findJob(w http.ResponseWriter, r *http.Request) {
	jobName := chi.URLParam(r, "job_name")
	projectName := chi.URLParam(r, "project_name")

	jobSpecResp, err := h.service.FindJobSpec(r.Context(), jobName, projectName)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, jobSpecResp)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")

	listResp, err := h.service.ListJobs(r.Context(), projectName)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, listResp)
}
