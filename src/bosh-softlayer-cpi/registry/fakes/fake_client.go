package fakes

import (
	"bosh-softlayer-cpi/registry"
)

// FakeClient represents a Fake BOSH Registry Client.
type FakeClient struct {
	DeleteCalled bool
	DeleteErr    error

	FetchCalled   bool
	FetchErr      error
	FetchSettings registry.AgentSettings

	UpdateCalled   bool
	UpdateErr      error
	UpdateSettings registry.AgentSettings

	IsExistCalled bool
	Exist         bool
	IsExistErr    error
}

// Delete deletes the instance settings for a given instance ID.
func (c *FakeClient) Delete(instanceID string) error {
	c.DeleteCalled = true
	return c.DeleteErr
}

// Fetch gets the agent settings for a given instance ID.
func (c *FakeClient) Fetch(instanceID string) (registry.AgentSettings, error) {
	c.FetchCalled = true
	return c.FetchSettings, c.FetchErr
}

// Update updates the agent settings for a given instance ID.
func (c *FakeClient) Update(instanceID string, agentSettings registry.AgentSettings) error {
	c.UpdateCalled = true
	c.UpdateSettings = agentSettings
	return c.UpdateErr
}

// IsExist indicates whether the agent settings for a given instance ID exists w/i empty.
func (c *FakeClient) IsExist(instanceID string) (bool, error) {
	c.IsExistCalled = true
	return c.Exist, c.IsExistErr
}
