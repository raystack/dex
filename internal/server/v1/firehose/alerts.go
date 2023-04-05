package firehose

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/utils"
	alertsv1 "github.com/goto/dex/internal/server/v1/alert"
	"github.com/goto/dex/internal/server/v1/project"
	"github.com/goto/dex/pkg/errors"
)

const firehoseOutputReleaseNameKey = "release_name"

var suppliedAlertVariableNames = []string{"name", "team", "entity"}

func (api *firehoseAPI) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)
	prjSlug := projectSlugFromURN(urn)

	prj, err := project.GetProject(r.Context(), prjSlug, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	firehoseDef, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	name, err := getFirehoseReleaseName(*firehoseDef)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	alerts, err := api.AlertSvc.ListAlerts(r.Context(), prj.GetSlug(), name)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK,
		utils.ListResponse[alertsv1.Alert]{Items: alerts})
}

func (api *firehoseAPI) handleGetAlertPolicy(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)
	prjSlug := projectSlugFromURN(urn)

	prj, err := project.GetProject(r.Context(), prjSlug, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	firehoseDef, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	releaseName, err := getFirehoseReleaseName(*firehoseDef)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	policy, err := api.AlertSvc.GetAlertPolicy(r.Context(), prj.GetSlug(), releaseName)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	policy.Rules = alertsv1.RemoveSuppliedVariablesFromRules(policy.Rules, suppliedAlertVariableNames)

	utils.WriteJSON(w, http.StatusOK, policy)
}

func (api *firehoseAPI) handleUpsertAlertPolicy(w http.ResponseWriter, r *http.Request) {
	var policyDef alertsv1.Policy
	if err := utils.ReadJSON(r, &policyDef); err != nil {
		utils.WriteErr(w, err)
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	prjSlug := projectSlugFromURN(urn)

	firehoseDef, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	name, err := getFirehoseReleaseName(*firehoseDef)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	group := firehoseDef.Group.String()

	prj, err := project.GetProject(r.Context(), prjSlug, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	entity, err := api.AlertSvc.GetProjectDataSource(r.Context(), prj.GetSlug())
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	policyDef.Rules = alertsv1.AddSuppliedVariablesFromRules(policyDef.Rules, map[string]string{
		"team":   group,
		"name":   name,
		"entity": entity,
	})
	policyDef.Resource = name

	alertPolicy, err := api.AlertSvc.UpsertAlertPolicy(r.Context(), prj.GetSlug(), policyDef)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, alertPolicy)
}

func (api *firehoseAPI) stopAlerts(ctx context.Context, firehoseDef models.Firehose, prjSlug string) error {
	name, err := getFirehoseReleaseName(firehoseDef)
	if err != nil {
		return err
	}

	policy := alertsv1.Policy{
		Resource: name,
		Rules:    nil,
	}

	_, err = api.AlertSvc.UpsertAlertPolicy(ctx, prjSlug, policy)
	if errors.Is(err, errors.ErrNotFound) {
		err = nil
	}
	return err
}

func getFirehoseReleaseName(firehoseDef models.Firehose) (string, error) {
	errFail := errors.ErrInternal.WithMsgf("failed to parse release name")

	if firehoseDef.State == nil {
		return "", errFail.WithCausef("nil state")
	}

	output, ok := firehoseDef.State.Output.(map[string]any)
	if !ok {
		return "", errFail.WithCausef("output is not a map")
	}

	s, ok := output[firehoseOutputReleaseNameKey].(string)
	if !ok {
		return "", errFail.WithCausef("release name key not found")
	}
	return s, nil
}
