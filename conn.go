package amqputils

import "github.com/streadway/amqp"

// SubscribeFunc function to handle an incoming message
type SubscribeFunc func(amqp.Delivery) []byte

// CreateConnection channel and its respective close function
func CreateConnection(url string) (*amqp.Channel, func(), error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	close := func() {
		conn.Close()
		ch.Close()
	}
	return ch, close, nil
}

// CreateQueue in the amqp server
func CreateQueue(ch *amqp.Channel, queueName string) (*amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return nil, err
	}

	return &q, nil
}

// Subscribe to a queue and handle the messages
func Subscribe(ch *amqp.Channel, q *amqp.Queue, do SubscribeFunc) error {
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil
	}

	for d := range msgs {
		msg := do(d)
		if msg != nil && d.ReplyTo != "" && d.CorrelationId != "" {
			ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          msg,
				})
		}
	}
	return nil
}
