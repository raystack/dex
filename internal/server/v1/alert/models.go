package alert

type SubscriptionForm struct {
	UserID             string
	AlertSeverity      AlertSeverity
	ProjectID          string
	ChannelCriticality ChannelCriticality
	GroupID            string
	ResourceID         string
	ResourceType       string
}
