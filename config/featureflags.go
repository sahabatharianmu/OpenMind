package config

// FeatureFlagConfig holds feature flag service configuration
type FeatureFlagConfig struct {
	// Provider is the feature flag service to use ("launchdarkly" or "local")
	Provider string `mapstructure:"provider"`

	// SDKKey is the server-side SDK key for LaunchDarkly
	SDKKey string `mapstructure:"sdk_key"`
}

// DefaultFeatureFlagConfig returns default feature flag configuration
func DefaultFeatureFlagConfig() *FeatureFlagConfig {
	return &FeatureFlagConfig{
		Provider: "local",
		SDKKey:   "",
	}
}
