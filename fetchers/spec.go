package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// InstanceSpecFetcher holds BOSH instance metadata from the spec.json
type InstanceSpecFetcher struct {
	specPath string // /var/vcap/bosh/spec.json
}

type InstanceSpec struct {
	Deployment string `json:"deployment"`
	Name       string `json:"name"`
	Index      int    `json:"index"`
	ID         string `json:"id"`
	AZ         string `json:"az"`
}

func NewInstanceSpecFetcher(specPath string) *InstanceSpecFetcher {
	return &InstanceSpecFetcher{specPath: specPath}
}

func (m *InstanceSpecFetcher) Fetch(_ context.Context) (*InstanceSpec, error) {
	var spec InstanceSpec
	data, err := os.ReadFile(m.specPath)
	if err != nil {
		return &spec, fmt.Errorf("cannot read instance spec file '%s', error: %v", m.specPath, err)
	}
	return m.FetchData(data)
}

func (m *InstanceSpecFetcher) FetchData(data []byte) (*InstanceSpec, error) {
	var spec InstanceSpec
	err := json.Unmarshal(data, &spec)
	if err != nil {
		return &spec, fmt.Errorf("cannot parse instance spec file '%s', error: %v", m.specPath, err)
	}
	return &spec, nil
}
