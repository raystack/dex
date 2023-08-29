package alert

import sirenv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/siren/v1beta1"

type SubscriptionForm struct {
	UserID             string
	AlertSeverity      AlertSeverity
	ProjectID          string
	ChannelCriticality ChannelCriticality
	GroupID            string
	ResourceID         string
	ResourceType       string
}

type AlertChannelForm struct {
	ChannelCriticality  ChannelCriticality
	ChannelName         string
	ChannelType         string
	PagerdutyServiceKey string
}

type SirenReceivers []*sirenv1beta1.Receiver

func (receivers SirenReceivers) Find(receiverType, severity string) *sirenv1beta1.Receiver {
	for _, r := range receivers {
		if r.Type == receiverType && r.Labels[sirenReceiverLabelKeySeverity] == severity {
			return r
		}
	}
	return nil
}
