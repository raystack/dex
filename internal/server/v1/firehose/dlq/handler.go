package dlq

import (
	"net/http"

	"github.com/goto/dex/internal/server/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) listDlq(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"dlq_list": []interface{}{},
	})
}

func (h *Handler) listDlqJobs(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"dlq_jobs": []interface{}{},
	})
}

func (h *Handler) createDlqJob(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"dlq_job": nil,
	})
}

func (h *Handler) getDlqJob(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"dlq_job": nil,
	})
}
