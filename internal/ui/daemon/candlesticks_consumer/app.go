package candlesticks_consumer

import (
	"context"
	"github.com/streadway/amqp"
	_ "github.com/streadway/amqp"
	"log"
)

type Application struct {
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Run(ctx context.Context) {
	conn, err := amqp.Dial("amqp://admin:crypto@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("candlesticks-queue", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	_ = q

	// Публикация сообщения
	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte("Hello, RabbitMQ!"),
	})
	if err != nil {
		log.Fatal(err)
	}
	d, err := ch.Consume("candlesticks-queue", "candlesticks-consumer", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-ctx.Done():
			log.Println("Context done")
			return
		case d, ok := <-d:
			if !ok {
				log.Println("Channel closed")
				return
			}

			log.Printf("Received a message: %s", d.Body)
			if err := d.Ack(false); err != nil {
				log.Printf("Failed to acknowledge message: %s", err)
			}
		}
	}

	log.Println("Message published!")
}
