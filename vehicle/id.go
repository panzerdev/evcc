package vehicle

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/vehicle/id"
	"github.com/evcc-io/evcc/vehicle/vw"
)

// https://github.com/TA2k/ioBroker.vw-connect

// ID is an api.Vehicle implementation for ID cars
type ID struct {
	*Embed
	*id.Provider // provides the api implementations
}

func init() {
	registry.Add("id", NewIDFromConfig, defaults().WithTimeout())
}

// NewIDFromConfig creates a new vehicle
func NewIDFromConfig(other map[string]interface{}) (api.Vehicle, error) {
	cc := defaults().WithTimeout()

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}

	v := &ID{
		Embed: &cc.Embed,
	}

	log := util.NewLogger("id").Redact(cc.User, cc.Password, cc.VIN)
	identity := vw.NewIdentity(log)

	query := url.Values(map[string][]string{
		"response_type": {"code id_token token"},
		"client_id":     {"a24fba63-34b3-4d43-b181-942111e6bda8@apps_vw-dilab_com"},
		"redirect_uri":  {"weconnect://authenticated"},
		"scope":         {"openid profile badge cars dealers vin"},
	})

	ts := id.NewTokenSource(log, identity, query, cc.User, cc.Password)
	err := identity.Login(ts)
	if err != nil {
		return v, fmt.Errorf("login failed: %w", err)
	}

	api := id.NewAPI(log, identity)
	api.Client.Timeout = cc.Timeout

	if cc.VIN == "" {
		cc.VIN, err = findVehicle(api.Vehicles())
		if err == nil {
			log.DEBUG.Printf("found vehicle: %v", cc.VIN)
		}
	}

	v.Provider = id.NewProvider(api, strings.ToUpper(cc.VIN), cc.Cache)

	return v, err
}
