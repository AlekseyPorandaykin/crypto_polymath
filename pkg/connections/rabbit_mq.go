package connections

import (
	"fmt"
	"github.com/streadway/amqp"
)

var rabbitMqConnections = make(map[string]*amqp.Connection)

type RabbitConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Queue    string `yaml:"queue"`
	Exchange string `yaml:"exchange"`
	Consumer string `yaml:"consumer"`
}

func CreateRabbitConnection(c RabbitConfig) (*amqp.Connection, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", c.Username, c.Password, c.Host, c.Port)
	conn, _ := rabbitMqConnections[dsn]
	if conn == nil {
		var err error
		conn, err = amqp.Dial(dsn)
		if err != nil {
			return nil, err
		}
		rabbitMqConnections[dsn] = conn
	}
	return conn, nil
}

func CreateRabbitDelivery(c RabbitConfig) (<-chan amqp.Delivery, error) {
	conn, err := CreateRabbitConnection(c)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	d, err := ch.Consume(c.Queue, c.Consumer, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func CreateRabbitPublisher(c RabbitConfig) (*amqp.Channel, error) {
	conn, err := CreateRabbitConnection(c)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	_, errD := ch.QueueDeclare(c.Queue, false, false, false, false, nil)
	if errD != nil {
		return nil, errD
	}
	return ch, nil
}
