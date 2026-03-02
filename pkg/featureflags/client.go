package featureflags

import (
	"context"
	"time"

	ldcontext "github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	ld "github.com/launchdarkly/go-server-sdk/v7"
	"github.com/launchdarkly/go-server-sdk/v7/ldcomponents"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

// Client wraps the LaunchDarkly SDK client
type Client struct {
	ld       *ld.LDClient
	logger   logger.Logger
	provider string
}

// UserContext holds user attributes for flag evaluation
type UserContext struct {
	UserID string
	Email  string
	OrgID  string
	Role   string
}

// New creates a new feature flag client
// If provider is "local" or sdkKey is empty, returns a no-op client
func New(provider, sdkKey string, log logger.Logger) (*Client, error) {
	c := &Client{
		logger:   log,
		provider: provider,
	}

	if provider != "launchdarkly" || sdkKey == "" {
		log.Info("Feature flags running in local mode (all flags default)")
		return c, nil
	}

	config := ld.Config{
		Events: ldcomponents.NoEvents(),
	}

	client, err := ld.MakeCustomClient(sdkKey, config, 10*time.Second) //nolint:mnd
	if err != nil {
		log.Error("Failed to initialize LaunchDarkly", zap.Error(err))
		return c, err
	}

	c.ld = client
	log.Info("LaunchDarkly client initialized successfully")
	return c, nil
}

// buildContext creates an LD evaluation context from user attributes
func (c *Client) buildContext(user UserContext) ldcontext.Context {
	builder := ldcontext.NewBuilder(user.UserID)
	builder.Kind("user")

	if user.Email != "" {
		builder.SetString("email", user.Email)
	}
	if user.OrgID != "" {
		builder.SetString("orgId", user.OrgID)
	}
	if user.Role != "" {
		builder.SetString("role", user.Role)
	}

	ctx, _ := builder.TryBuild()
	return ctx
}

// IsEnabled evaluates a boolean feature flag
func (c *Client) IsEnabled(flagKey string, user UserContext) bool {
	if c.ld == nil {
		return false // local mode: all flags off by default
	}

	ctx := c.buildContext(user)
	value, err := c.ld.BoolVariation(flagKey, ctx, false)
	if err != nil {
		c.logger.Warn("Flag evaluation error",
			zap.String("flag", flagKey),
			zap.Error(err),
		)
		return false
	}

	return value
}

// GetStringVariation evaluates a string feature flag (useful for A/B variants)
func (c *Client) GetStringVariation(flagKey string, user UserContext, defaultVal string) string {
	if c.ld == nil {
		return defaultVal
	}

	ctx := c.buildContext(user)
	value, err := c.ld.StringVariation(flagKey, ctx, defaultVal)
	if err != nil {
		c.logger.Warn("Flag evaluation error",
			zap.String("flag", flagKey),
			zap.Error(err),
		)
		return defaultVal
	}

	return value
}

// Close shuts down the LaunchDarkly client gracefully
func (c *Client) Close() error {
	if c.ld != nil {
		c.logger.Info("Shutting down LaunchDarkly client")
		return c.ld.Close()
	}
	return nil
}

// IsReady returns true if the client is initialized and connected
func (c *Client) IsReady() bool {
	if c.ld == nil {
		return c.provider == "local"
	}
	return c.ld.Initialized()
}

// Provider returns the current provider name
func (c *Client) Provider() string {
	return c.provider
}

// Background returns a context for server-side flag evaluation without a user
func Background() context.Context {
	return context.Background()
}
