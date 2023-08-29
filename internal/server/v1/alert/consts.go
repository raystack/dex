package alert

// SIREN KEYS
const (
	groupMetadataKey        = "group_id"
	resourceIDMetadataKey   = "resource_id"
	resourceTypeMetadataKey = "resource_type"

	sirenReceiverLabelKeyOrg          = "org"
	sirenReceiverLabelKeyTeam         = "team"
	sirenReceiverLabelKeySeverity     = "severity"
	sirenReceiverConfigKeyChannelName = "channel_name"
	sirenReceiverConfigKeyServiceKey  = "service_key"
)

// SHIELD KEYS
const (
	shieldOrgMetadataKeySirenNamespaceID           = "siren_namespace_id"
	shieldOrgMetadataKeySirenParentSlackReceiverID = "siren_parent_slack_receiver_id"
)
