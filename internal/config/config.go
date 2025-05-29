package config

import (
	"github.com/AlekseyPorandaykin/go-template/pkg/connection"
	"github.com/spf13/viper"
)

type AppConf struct {
	DBConnection connection.DBConfig
}

func Create() AppConf {
	return AppConf{
		DBConnection: connection.DBConfig{
			Driver:             viper.GetString("db_connection.driver"),
			Username:           viper.GetString("db_connection.username"),
			Password:           viper.GetString("db_connection.password"),
			Host:               viper.GetString("db_connection.host"),
			Port:               viper.GetString("db_connection.port"),
			Database:           viper.GetString("db_connection.database"),
			PathToDB:           viper.GetString("db_connection.path_to_db"),
			SchemaName:         viper.GetString("db_connection.schema"),
			MaxOpenConnections: 5,
			MaxIdleConnections: 5,
		},
	}
}
