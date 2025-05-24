package pkg

import (
	"encoding/json"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ connection constants
const (
	TransferQueueName      = "transfer_queue"
	TransferExchangeName   = "transfer_exchange"
	TransferRoutingKey     = "transfer.request"
	ReconnectDelay         = 5 * time.Second
	ResendDelay            = 5 * time.Second
	ReconnectRetryAttempts = 10
)

// getRabbitMQURL returns the RabbitMQ URL from environment or default
func getRabbitMQURL() string {
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@localhost:5672/"
}

// TransferMessage represents a transfer task to be processed
type TransferMessage struct {
	TransferID  string  `json:"transfer_id"`
	SenderID    string  `json:"sender_id"`
	RecipientID string  `json:"recipient_id"`
	Amount      float64 `json:"amount"`
	Remarks     string  `json:"remarks"`
}

// RabbitMQ holds the connection and channel
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	isReady bool
}

// IsReady returns true if the RabbitMQ connection is ready
func (r *RabbitMQ) IsReady() bool {
	return r != nil && r.isReady
}

var GlobalRabbitMQ *RabbitMQ

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ() (*RabbitMQ, error) {
	rmq := &RabbitMQ{isReady: false}
	if err := rmq.Connect(); err != nil {
		return nil, err
	}
	return rmq, nil
}

// Connect establishes a connection to RabbitMQ
func (r *RabbitMQ) Connect() error {
	var err error

	// Connect to RabbitMQ using environment URL
	rabbitmqURL := getRabbitMQURL()
	log.Printf("Connecting to RabbitMQ at: %s", rabbitmqURL)
	r.conn, err = amqp.Dial(rabbitmqURL)
	if err != nil {
		return err
	}

	// Create channel
	r.channel, err = r.conn.Channel()
	if err != nil {
		r.conn.Close()
		return err
	}

	// Setup our topology
	err = r.setupTopology()
	if err != nil {
		r.Close()
		return err
	}

	r.isReady = true
	return nil
}

// setupTopology sets up exchanges and queues
func (r *RabbitMQ) setupTopology() error {
	// Declare exchange
	err := r.channel.ExchangeDeclare(
		TransferExchangeName, // name
		"direct",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return err
	}

	// Declare quorum queue with appropriate arguments
	args := make(amqp.Table)
	args["x-queue-type"] = "quorum"
	// Optional: set delivery limit to prevent infinite redelivery loops
	args["x-delivery-limit"] = 5
	// Optional: set additional quorum queue properties
	args["x-max-in-memory-length"] = 1000

	_, err = r.channel.QueueDeclare(
		TransferQueueName, // name
		true,              // durable (quorum queues are always durable)
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		args,              // arguments with queue type
	)
	if err != nil {
		return err
	}

	// Bind queue to exchange
	err = r.channel.QueueBind(
		TransferQueueName,    // queue name
		TransferRoutingKey,   // routing key
		TransferExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the channel and connection
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
	r.isReady = false
}

// PublishTransfer publishes a transfer message to the queue
func (r *RabbitMQ) PublishTransfer(msg TransferMessage) error {
	if !r.isReady {
		return nil // Just skip if RabbitMQ is not ready
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		TransferExchangeName, // exchange
		TransferRoutingKey,   // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make messages persistent
		},
	)
}

// ConsumeTransfers starts consuming transfer messages from the queue
func (r *RabbitMQ) ConsumeTransfers() (<-chan amqp.Delivery, error) {
	if !r.isReady {
		return nil, nil // Return nil if RabbitMQ is not ready
	}

	// Set prefetch count to limit the number of unacknowledged messages
	// This is important for quorum queues to control memory usage
	if err := r.channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		return nil, err
	}

	return r.channel.Consume(
		TransferQueueName, // queue
		"",                // consumer
		false,             // auto-ack (must be false for quorum queues to ensure proper acknowledgment)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
}

// InitRabbitMQ initializes the global RabbitMQ instance
func InitRabbitMQ() error {
	var err error
	GlobalRabbitMQ, err = NewRabbitMQ()
	if err != nil {
		log.Printf("WARNING: Failed to initialize RabbitMQ: %v", err)
		GlobalRabbitMQ = &RabbitMQ{isReady: false}
		return err
	}
	return nil
}
