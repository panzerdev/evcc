package vehicle

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/vehicle/vw"
)

// https://github.com/trocotronic/weconnect
// https://github.com/TA2k/ioBroker.vw-connect

// VW is an api.Vehicle implementation for VW cars
type VW struct {
	*Embed
	*vw.Provider // provides the api implementations
}

func init() {
	registry.Add("vw", NewVWFromConfig, defaults().WithTimeout())
}

// NewVWFromConfig creates a new vehicle
func NewVWFromConfig(other map[string]interface{}) (api.Vehicle, error) {
	cc := defaults().WithTimeout()

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}

	v := &VW{
		Embed: &cc.Embed,
	}

	log := util.NewLogger("vw").Redact(cc.User, cc.Password, cc.VIN)
	identity := vw.NewIdentity(log)

	query := url.Values(map[string][]string{
		"response_type": {"id_token token"},
		"client_id":     {"9496332b-ea03-4091-a224-8c746b885068@apps_vw-dilab_com"},
		"redirect_uri":  {"carnet://identity-kit/login"},
		"scope":         {"openid profile mbb"}, // cars birthdate nickname address phone
	})

	ts := vw.NewTokenSource(log, identity, "38761134-34d0-41f3-9a73-c4be88d7d337", query, cc.User, cc.Password)
	err := identity.Login(ts)
	if err != nil {
		return v, fmt.Errorf("login failed: %w", err)
	}

	api := vw.NewAPI(log, identity, "VW", "DE")
	api.Client.Timeout = cc.Timeout

	if cc.VIN == "" {
		cc.VIN, err = findVehicle(api.Vehicles())
		if err == nil {
			log.DEBUG.Printf("found vehicle: %v", cc.VIN)
		}
	}

	if err == nil {
		if err = api.HomeRegion(strings.ToUpper(cc.VIN)); err == nil {
			v.Provider = vw.NewProvider(api, strings.ToUpper(cc.VIN), cc.Cache)
		}
	}

	return v, err
}
