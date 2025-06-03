package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlekseyPorandaykin/go-template/pkg/system"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"sync"
	"sync/atomic"
)

var rabbitMqConnections = make(map[string]*amqp.Connection)

func CreateRabbitConnection(username, password, addr string) (*amqp.Connection, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s/", username, password, addr)
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

var ErrClosed = errors.New("channel closed")

type RabbitMQProducer[T any] struct {
	ch   *amqp.Channel
	conf rabbitMQConfig
}

type rabbitMQConfig struct {
	// required
	QueueName string

	// options
	Durable     bool
	AutoDelete  bool
	Exclusive   bool
	NoWait      bool
	Mandatory   bool
	Immediate   bool
	ContentType string
	Exchange    string
	AutoAck     bool
	NoLocal     bool
	AskMultiple bool
	BatchSize   int
}

type RabbitMQOption func(*rabbitMQConfig)

func WithRabbitMQDurable(durable bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.Durable = durable
	}
}
func WithRabbitMQAutoDelete(autoDelete bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.AutoDelete = autoDelete
	}
}
func WithRabbitMQExclusive(exclusive bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.Exclusive = exclusive
	}
}
func WithRabbitMQNoWait(noWait bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.NoWait = noWait
	}
}
func WithRabbitMQMandatory(mandatory bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.Mandatory = mandatory
	}
}
func WithRabbitMQContentType(contentType string) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.ContentType = contentType
	}
}
func WithRabbitMQExchange(exchange string) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.Exchange = exchange
	}
}
func WithRabbitMQAutoAck(autoAck bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.AutoAck = autoAck
	}
}
func WithRabbitMQNoLocal(noLocal bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.NoLocal = noLocal
	}
}
func WithRabbitMQBatchSize(batchSize int) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.BatchSize = batchSize
	}
}
func WithRabbitMQAskMultiple(askMultiple bool) RabbitMQOption {
	return func(config *rabbitMQConfig) {
		config.AskMultiple = askMultiple
	}
}
func NewRabbitMQProducer[T any](conn *amqp.Connection, queueName string, options ...RabbitMQOption) (*RabbitMQProducer[T], error) {
	conf := &rabbitMQConfig{
		ContentType: "application/json",
	}
	for _, opt := range options {
		opt(conf)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a channel")
	}
	q, err := ch.QueueDeclare(queueName, conf.Durable, conf.AutoDelete, conf.Exclusive, conf.NoWait, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to declare a queue")
	}
	if q.Name == "" {
		return nil, errors.New("queue name is incorrect")
	}
	conf.QueueName = q.Name

	return &RabbitMQProducer[T]{ch: ch, conf: *conf}, nil
}

func (p *RabbitMQProducer[T]) publish(message T) error {
	body, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}
	return p.ch.Publish(p.conf.Exchange, p.conf.QueueName, p.conf.Mandatory, p.conf.Immediate, amqp.Publishing{
		ContentType: p.conf.ContentType,
		Body:        body,
	})
}

func (p *RabbitMQProducer[T]) Publish(messages ...T) error {
	for i := range messages {
		if err := p.publish(messages[i]); err != nil {
			return errors.Wrapf(err, "failed to publish message %d", i)
		}
	}
	return nil
}

func (p *RabbitMQProducer[T]) Close() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
}

type RabbitMQConsumer[T any] struct {
	ch         *amqp.Channel
	conf       rabbitMQConfig
	deliveryCh <-chan amqp.Delivery

	msgMu    sync.Mutex
	msgCh    chan T
	errCh    chan error
	isClosed atomic.Bool
}

func NewRabbitMQConsumer[T any](
	conn *amqp.Connection, queueName, consumerName string, options ...RabbitMQOption,
) (*RabbitMQConsumer[T], error) {
	conf := &rabbitMQConfig{
		ContentType: "application/json",
		BatchSize:   100,
	}
	for _, opt := range options {
		opt(conf)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a channel")
	}
	_, errDeclare := ch.QueueDeclare(queueName, conf.Durable, conf.AutoDelete, conf.Exclusive, conf.NoWait, nil)
	_ = errDeclare
	d, err := ch.Consume(queueName, consumerName, conf.AutoAck, conf.Exclusive, conf.NoLocal, conf.NoWait, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to consume messages")
	}
	c := &RabbitMQConsumer[T]{
		errCh:      make(chan error, 1),
		msgCh:      make(chan T, conf.BatchSize),
		deliveryCh: d,
	}
	system.Go(func() {
		c.Listen()
	})
	return c, nil
}

func (c *RabbitMQConsumer[T]) Listen() {
	for {
		if c.isClosed.Load() {
			return
		}
		select {
		case d, ok := <-c.deliveryCh:
			if !ok {
				return
			}
			var message T
			if err := json.Unmarshal(d.Body, &message); err != nil {
				c.errCh <- errors.Wrap(err, "failed to unmarshal message")
				continue
			}
			c.msgCh <- message
			if err := d.Ack(c.conf.AskMultiple); err != nil {
				c.errCh <- errors.Wrap(err, "failed to ack message")
				continue
			}
		}
	}
}
func (c *RabbitMQConsumer[T]) Consume(ctx context.Context, handler func(message T) error) error {
	for {
		if c.isClosed.Load() {
			return nil
		}
		select {
		case d, ok := <-c.deliveryCh:
			if !ok {
				return nil
			}
			var message T
			if err := json.Unmarshal(d.Body, &message); err != nil {
				return errors.Wrap(err, "failed to unmarshal message")
			}
			if err := handler(message); err != nil {
				return errors.Wrap(err, "handler error")
			}
			if err := d.Ack(c.conf.AskMultiple); err != nil {
				return errors.Wrap(err, "failed to ack message")
			}
		}
	}
}

func (c *RabbitMQConsumer[T]) Receive() (*T, error) {
	for {
		if c.isClosed.Load() {
			return nil, ErrClosed
		}
		select {
		case err := <-c.errCh:
			if err != nil {
				return nil, err
			}
		case m, ok := <-c.msgCh:
			if !ok {
				continue
			}
			return &m, nil
		}
	}
}

func (c *RabbitMQConsumer[T]) Close() {
	c.isClosed.Store(true)
	close(c.msgCh)
	close(c.errCh)
	if c.ch != nil {
		_ = c.ch.Close()
	}
}
