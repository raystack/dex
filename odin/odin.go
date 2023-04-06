package odin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/goto/dex/pkg/errors"
)

type odinStream struct {
	Urn           string `json:"urn"`
	Name          string `json:"name"`
	Entity        string `json:"entity"`
	Organization  string `json:"organization"`
	Landscape     string `json:"landscape"`
	Environment   string `json:"environment"`
	Type          string `json:"type"`
	AdvertiseMode struct {
		Host    string `json:"host"`
		Address string `json:"address"`
	} `json:"advertise_mode"`
	Brokers   []broker `json:"brokers"`
	Created   string   `json:"created"`
	Updated   string   `json:"updated"`
	ProjectID string   `json:"projectID"`
	URL       string   `json:"url"`
	ID        string   `json:"id"`
}

type broker struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	Address string `json:"address"`
}

func GetOdinStream(ctx context.Context, odinAddr, urn string) (string, error) {
	url := fmt.Sprintf("%s/streams/%s", odinAddr, urn)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", errors.ErrNotFound.WithMsgf("stream '%s' not found", urn)
		}
		return "", errors.ErrInternal.
			WithMsgf("failed to resolve stream").
			WithCausef("unexpected status code: %d", resp.StatusCode)
	}

	var odinResp odinStream
	err = json.NewDecoder(resp.Body).Decode(&odinResp)
	if err != nil {
		return "", err
	}

	return odinResp.URL, nil
}
