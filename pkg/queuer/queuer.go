package queuer

import (
	"errors"
	"log"

	"github.com/streadway/amqp"
)

var (
	errAMQPconnetction  = errors.New("Can not connect to AMQP")
	errOpenChannel      = errors.New("Failed to open a channel")
	errAMQPDeclare      = errors.New("Failed to declare a queue")
	errRegisterConsumer = errors.New("Failed to register a consumer")
)

type Queuer struct {
	amqpCon   *amqp.Connection
	queue     amqp.Queue
	messageCh chan amqp.Delivery
	stopCh    chan bool
	// tasks chan amqp.
}

func NewQueuer() *Queuer {
	q := &Queuer{
		messageCh: make(chan amqp.Delivery, 1),
		stopCh:    make(chan bool),
	}
	return q
}

func (q *Queuer) initAMQPCon() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return errAMQPconnetction
	}
	q.amqpCon = conn

	ch, err := conn.Channel()
	if err != nil {
		return errOpenChannel
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"grader", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return errAMQPDeclare
	}
	q.queue = queue
	// Don't forget
	// defer conn.Close()
	return nil
}

func (q *Queuer) Run() error {
	ch, err := q.amqpCon.Channel()
	if err != nil {
		return errOpenChannel
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"grader", // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return errRegisterConsumer
	}
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			q.messageCh <- d
		}
	}()

	<-q.stopCh
	return nil
}

func (q *Queuer) Stop() {
	q.stopCh <- true
}
