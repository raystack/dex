package firehose

import (
	"context"
	"fmt"
	"net/http"

	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	"github.com/go-chi/chi/v5"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/utils"
	alertsv1 "github.com/goto/dex/internal/server/v1/alert"
	"github.com/goto/dex/internal/server/v1/project"
	"github.com/goto/dex/pkg/errors"
)

const (
	resourceTag = "firehose"
)

var suppliedAlertVariableNames = []string{"name", "team", "entity"}

func (api *firehoseAPI) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)

	firehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	prj, err := project.GetProject(r.Context(), firehose.Project, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	alerts, err := api.AlertSvc.ListAlerts(r.Context(), prj.GetSlug(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK,
		utils.ListResponse[alertsv1.Alert]{Items: alerts})
}

func (api *firehoseAPI) handleGetAlertPolicy(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)

	firehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	prj, err := project.GetProject(r.Context(), firehose.Project, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	policy, err := api.AlertSvc.GetAlertPolicy(r.Context(), prj.GetSlug(), urn, resourceTag)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	policy.Rules = alertsv1.RemoveSuppliedVariablesFromRules(policy.Rules, suppliedAlertVariableNames)

	utils.WriteJSON(w, http.StatusOK, policy)
}

func (api *firehoseAPI) handleUpsertAlertPolicy(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)

	var policyDef alertsv1.Policy
	if err := utils.ReadJSON(r, &policyDef); err != nil {
		utils.WriteErr(w, err)
		return
	}

	firehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	resp, err := api.Shield.GetGroup(r.Context(), &shieldv1beta1.GetGroupRequest{
		Id: firehose.Group.String(),
	})
	if err != nil {
		utils.WriteErr(w, fmt.Errorf("could not find group: %w", err))
		return
	}
	group := resp.Group

	prj, err := project.GetProject(r.Context(), firehose.Project, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	projectSlug := prj.GetSlug()

	entity, err := api.AlertSvc.GetProjectDataSource(r.Context(), projectSlug)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	policyDef.Rules = alertsv1.AddSuppliedVariablesFromRules(policyDef.Rules, map[string]string{
		"team":   group.Slug,
		"name":   urn,
		"entity": entity,
	})
	policyDef.Resource = urn

	alertPolicy, err := api.AlertSvc.UpsertAlertPolicy(r.Context(), projectSlug, policyDef)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, alertPolicy)
}

func (api *firehoseAPI) stopAlerts(ctx context.Context, firehose models.Firehose, prjSlug string) error {
	policy := alertsv1.Policy{
		Resource: firehose.Urn,
		Rules:    nil,
	}

	_, err := api.AlertSvc.UpsertAlertPolicy(ctx, prjSlug, policy)
	if errors.Is(err, errors.ErrNotFound) {
		err = nil
	}
	return err
}
