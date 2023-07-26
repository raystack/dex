package firehose

import (
	"github.com/go-openapi/strfmt"
)

func sanitizeFirehoseStopTime(val *strfmt.DateTime) *strfmt.DateTime {
	if val != nil && val.Equal(strfmt.NewDateTime()) {
		return nil
	}

	return val
}
