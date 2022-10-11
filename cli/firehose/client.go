package firehose

import (
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/odpf/dex/generated/client"
)

func initClient() *client.DexAPI {
	r := httptransport.New(client.DefaultHost, client.DefaultBasePath, client.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(os.Getenv("API_ACCESS_TOKEN"))
	return client.New(r, strfmt.Default)
}
