package alert

import (
	"fmt"

	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"
	"github.com/go-openapi/strfmt"

	"github.com/goto/dex/generated/models"
)

func MapToSubscriptionList(sirenSubs []*sirenv1beta1.Subscription) []models.Subscription {
	subs := []models.Subscription{}
	for _, sirenSub := range sirenSubs {
		subs = append(subs, MapToSubscription(sirenSub))
	}
	return subs
}

func MapToSubscription(sirenSub *sirenv1beta1.Subscription) models.Subscription {
	receivers := mapSirenReceiversToModels(sirenSub.Receivers)

	return models.Subscription{
		ID:        fmt.Sprintf("%d", sirenSub.Id),
		Urn:       sirenSub.Urn,
		CreatedAt: strfmt.DateTime(sirenSub.CreatedAt.AsTime()),
		CreatedBy: sirenSub.CreatedBy,
		Match:     sirenSub.Match,
		Metadata:  sirenSub.Metadata,
		Namespace: fmt.Sprintf("%d", sirenSub.Namespace),
		Receivers: receivers,
		UpdatedAt: strfmt.DateTime(sirenSub.UpdatedAt.AsTime()),
		UpdatedBy: sirenSub.UpdatedBy,
	}
}

func mapSirenReceiversToModels(sirenRecvs []*sirenv1beta1.ReceiverMetadata) []*models.SubscriptionReceiversItems0 {
	recvs := []*models.SubscriptionReceiversItems0{}
	for _, sirenRecv := range sirenRecvs {
		recvs = append(recvs, &models.SubscriptionReceiversItems0{
			ID:            fmt.Sprintf("%d", sirenRecv.Id),
			Configuration: sirenRecv.Configuration,
		})
	}

	return recvs
}
