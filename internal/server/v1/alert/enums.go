package alert

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "INFO"
	AlertSeverityWarning  AlertSeverity = "WARNING"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

type ChannelCriticality string

const (
	ChannelCriticalityInfo     ChannelCriticality = "INFO"
	ChannelCriticalityWarning  ChannelCriticality = "WARNING"
	ChannelCriticalityCritical ChannelCriticality = "CRITICAL"
)