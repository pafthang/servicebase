package migrations

import (
	"encoding/json"

	"github.com/pafthang/servicebase/daos"
	settingsmodels "github.com/pafthang/servicebase/services/settings/models"
	"github.com/pocketbase/dbx"
)

// register001Init creates the `_params` table for the settings module and
// bootstraps the default application settings payload when missing. It also
// normalizes legacy unencrypted settings payloads so newly added settings
// fields are persisted explicitly in the stored JSON.
func register001Init(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		if _, err := db.NewQuery(`
			CREATE TABLE IF NOT EXISTS {{_params}} (
				[[id]]      TEXT PRIMARY KEY NOT NULL,
				[[key]]     TEXT UNIQUE NOT NULL,
				[[value]]   JSON DEFAULT NULL,
				[[created]] TEXT DEFAULT "" NOT NULL,
				[[updated]] TEXT DEFAULT "" NOT NULL
			);
		`).Execute(); err != nil {
			return err
		}

		dao := daos.New(db)

		param, err := dao.FindParamByKey(settingsmodels.ParamAppSettings)
		if err != nil || param == nil || len(param.Value) == 0 {
			return dao.SaveSettings(settingsmodels.NewSettings())
		}

		settings := settingsmodels.NewSettings()
		if err := json.Unmarshal(param.Value, settings); err != nil {
			// Encrypted or otherwise non-plain settings payloads are left as-is.
			return nil
		}

		encoded, err := json.Marshal(settings)
		if err != nil {
			return err
		}

		param.Value = encoded
		return dao.Save(param)
	}, func(db dbx.Builder) error {
		return nil
	}, "services_settings_001_init.go")
}
