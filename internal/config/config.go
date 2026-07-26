package config

import (
	"github.com/AlekseyPorandaykin/go-kit/pkg/connection"
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
			MaxOpenConnections: viper.GetInt("db_connection.max_open_connections"),
			MaxIdleConnections: viper.GetInt("db_connection.max_idle_connections"),
		},
	}
}
