package charger

import (
	"context"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/charger/config"
)

var registry = config.Registry

func RegisterChargerCtx(name string, charger func(context.Context, map[string]any) (api.Charger, error)) {
	registry.AddCtx(name, charger)
}

func RegisterCharger(name string, charger func(map[string]any) (api.Charger, error)) {
	registry.Add(name, charger)
}

// Types returns the list of types
func Types() []string {
	return registry.Types()
}

// NewFromConfig creates charger from configuration
func NewFromConfig(ctx context.Context, typ string, other map[string]interface{}) (api.Charger, error) {
	return config.NewFromConfig(ctx, typ, other)
}
