package queuer

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
)

var (
	errAMQPconnetction  = errors.New("Can not connect to AMQP")
	errOpenChannel      = errors.New("Failed to open a channel")
	errAMQPDeclare      = errors.New("Failed to declare a queue")
	errRegisterConsumer = errors.New("Failed to register a consumer")
)

var (
	graderURL      = "http://127.0.0.1:8081"
	amqpDSN        = "amqp://guest:guest@localhost:5672/"
	defaultQueue   = "grader"
	defaultTimeout = 300
)

type Queuer struct {
	amqpCon   *amqp.Connection
	queue     amqp.Queue
	messageCh chan amqp.Delivery
	stopCh    chan bool
	// tasks chan amqp.
}

type Task struct {
	ID           int    `json:"id"`
	Status       int    `json:"status"`
	Course       string `json:"course"`
	Name         string `json:"name"`
	Filename     string `json:"filename"`
	S3BucketName string `json:"bucket"`
	UserID       int    `json:"user_id"`
}

type GraderTask struct {
	Name    string `json:"name"`
	Timeout int    `json:"timeout"`
}

func NewQueuer() *Queuer {
	q := &Queuer{
		messageCh: make(chan amqp.Delivery, 8),
		stopCh:    make(chan bool),
	}
	q.initAMQPCon()
	return q
}

func (q *Queuer) initAMQPCon() error {
	conn, err := amqp.Dial(amqpDSN)
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
		defaultQueue, // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
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
		defaultQueue, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return errRegisterConsumer
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// q.messageCh <- d
			q.SendTask(d.Body)
			d.Ack(false)
		}
	}()

	log.Println("Queuer Start")
	<-q.stopCh
	return nil
}

func (q *Queuer) Stop() {
	q.stopCh <- true
}

func (q *Queuer) SendTask(data []byte) {

	fmt.Println("GRADER URL:>", graderURL)

	req, err := http.NewRequest("POST", graderURL, bytes.NewBuffer(data))
	if err != nil {
		log.Println("NewRequest ERR:", err)
	}
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
