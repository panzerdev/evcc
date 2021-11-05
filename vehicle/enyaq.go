package vehicle

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/vehicle/skoda"
	"github.com/evcc-io/evcc/vehicle/vw"
)

// https://github.com/lendy007/skodaconnect

// Enyaq is an api.Vehicle implementation for Skoda Enyaq cars
type Enyaq struct {
	*Embed
	*skoda.Provider // provides the api implementations
}

func init() {
	registry.Add("enyaq", NewEnyaqFromConfig, defaults().WithTimeout())
}

// NewEnyaqFromConfig creates a new vehicle
func NewEnyaqFromConfig(other map[string]interface{}) (api.Vehicle, error) {
	cc := defaults().WithTimeout()

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}

	v := &Enyaq{
		Embed: &cc.Embed,
	}

	var err error
	log := util.NewLogger("enyaq").Redact(cc.User, cc.Password, cc.VIN)

	if cc.VIN == "" {
		identity := vw.NewIdentity(log)

		// Skoda native api
		query := url.Values(map[string][]string{
			"response_type": {"code id_token"},
			"redirect_uri":  {"skodaconnect://oidc.login/"},
			"client_id":     {"f9a2359a-b776-46d9-bd0c-db1904343117@apps_vw-dilab_com"},
			"scope":         {"openid profile mbb"},
		})

		ts := skoda.NewTokenSource(log, identity, query, cc.User, cc.Password)
		if err = identity.Login(ts); err != nil {
			return v, fmt.Errorf("login failed: %w", err)
		}

		api := skoda.NewAPI(log, identity)
		api.Client.Timeout = cc.Timeout

		cc.VIN, err = findVehicle(api.Vehicles())
		if err == nil {
			log.DEBUG.Printf("found vehicle: %v", cc.VIN)
		}
	}

	if err == nil {
		identity := vw.NewIdentity(log)

		// Skoda connect api
		query := url.Values(map[string][]string{
			"response_type": {"code id_token"},
			"redirect_uri":  {"skodaconnect://oidc.login/"},
			"client_id":     {"7f045eee-7003-4379-9968-9355ed2adb06@apps_vw-dilab_com"},
			"scope":         {"openid profile mbb"}, // phone address cars email birthdate badge dealers driversLicense
		})

		ts := skoda.NewTokenSource(log, identity, query, cc.User, cc.Password)
		if err = identity.Login(ts); err != nil {
			return v, fmt.Errorf("login failed: %w", err)
		}

		api := skoda.NewAPI(log, identity)
		api.Client.Timeout = cc.Timeout

		v.Provider = skoda.NewProvider(api, strings.ToUpper(cc.VIN), cc.Cache)
	}

	return v, err
}
