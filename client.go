package amqputils

import (
	"math/rand"
	"time"

	"github.com/streadway/amqp"
)

// Call a queue and receives the response
func Call(url, queueName string, info []byte) ([]byte, error) {
	ch, close, err := CreateConnection(url)
	if err != nil {
		return nil, err
	}
	defer close()
	return CallWithConn(ch, queueName, info)
}

// CallWithConn a queue and receives the response
func CallWithConn(ch *amqp.Channel, queueName string, info []byte) ([]byte, error) {
	q, err := CreateQueue(ch, queueName)
	if err != nil {
		return nil, err
	}

	qRec, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		return nil, err
	}

	corrID := randomString(32)
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode:  amqp.Persistent,
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       qRec.Name,
			Body:          info,
		})
	if err != nil {
		return nil, err
	}

	resp := make(chan []byte)
	go Subscribe(ch, &qRec, func(d amqp.Delivery) []byte {
		if corrID == d.CorrelationId {
			resp <- d.Body
		}
		return nil
	})

	select {
	case data := <-resp:
		return data, nil
	case <-time.NewTimer(5 * time.Second).C:
		return nil, ErrTimeout
	}
}

// Publish in a queue
func Publish(url, queueName string, info []byte) error {
	ch, close, err := CreateConnection(url)
	if err != nil {
		return err
	}
	defer close()
	return PublishWithConn(ch, queueName, info)
}

// PublishWithConn in a queue
func PublishWithConn(ch *amqp.Channel, queueName string, info []byte) error {
	q, err := CreateQueue(ch, queueName)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         info,
		})
	if err != nil {
		return err
	}
	return nil
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
