package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// RabbitMQ তে কানেক্ট করুন
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5673/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// চ্যানেল তৈরি করুন
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Exchange declare করুন (publisher এর মতোই)
	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Temporary queue তৈরি করুন (random name)
	q, err := ch.QueueDeclare(
		"",    // name empty = server generated
		false, // durable
		false, // delete when unused
		true,  // exclusive (connection close হলে queue delete)
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Bind queue to exchange
	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"logs", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	// Consumer register করুন
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		true,   // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// Message প্রসেস করুন
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
