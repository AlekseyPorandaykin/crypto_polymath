package config

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/database"
	"github.com/spf13/viper"
)

type AppConf struct {
	DBConnection database.Config
}

func Create() AppConf {
	return AppConf{
		DBConnection: database.Config{
			Driver:             viper.GetString("db_connection.driver"),
			PathToDB:           viper.GetString("db_connection.path_to_db"),
			MaxOpenConnections: 5,
			MaxIdleConnections: 5,
		},
	}
}
