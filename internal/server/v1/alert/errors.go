package alert

import "errors"

var (
	ErrInvalidAlertSeverity      = errors.New("invalid alert severity")
	ErrInvalidChannelCriticality = errors.New("invalid channel criticality")
	ErrSubscriptionNotFound      = errors.New("could not find subscription")
	ErrNoSirenReceiver           = errors.New("could not find siren's receiver")
	ErrMultipleSirenReceiver     = errors.New("multiple siren's receivers found")
	ErrInvalidSirenReceiver      = errors.New("invalid siren's receiver type")

	ErrNoShieldOrg                      = errors.New("could not find shield's organization")
	ErrNoShieldProject                  = errors.New("could not find shield's project")
	ErrNoShieldGroup                    = errors.New("could not find shield's group")
	ErrNoShieldParentSlackReceiver      = errors.New("could not find siren's parent slack receiver from shield metadata")
	ErrNoShieldSirenNamespace           = errors.New("could not find siren's namespace id from shield org metadata")
	ErrInvalidShieldSirenNamespace      = errors.New("invalid siren's namespace id from shield org metadata")
	ErrInvalidShieldParentSlackReceiver = errors.New("invalid siren's parent slack receiver from shield org metadata")
)
