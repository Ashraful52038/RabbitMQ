package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/rabbitmq/amqp091-go"
)

func main() {
    // Load configuration
    env := os.Getenv("APP_ENV")
    var config *RPCConfig
    
    switch env {
    case "production":
        config = ProductionRPCConfig()
        log.Println("Using PRODUCTION configuration")
    case "development":
        config = DevelopmentRPCConfig()
        log.Println("Using DEVELOPMENT configuration")
    default:
        config = DefaultRPCConfig()
        log.Println("Using DEFAULT configuration")
    }
    
    // Validate configuration
    if err := config.Validate(); err != nil {
        log.Fatalf("Invalid configuration: %v", err)
    }
    
    // Display configuration
    log.Println("========== RPC Server Configuration ==========")
    log.Printf("RabbitMQ: %s@%s:%d", 
        config.RabbitMQ.Username, 
        config.RabbitMQ.Host, 
        config.RabbitMQ.Port)
    log.Printf("Queue: %s", config.Queue.Name)
    log.Printf("QoS Prefetch: %d", config.QoS.PrefetchCount)
    log.Printf("Max Workers: %d", config.RPC.MaxWorkers)
    log.Println("==============================================")
    
    // Connect to RabbitMQ
    conn, err := amqp091.Dial(config.GetConnectionURL())
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()
    log.Println("Connected to RabbitMQ")
    
    // Create channel
    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open channel: %v", err)
    }
    defer ch.Close()
    
    // Apply QoS
    err = ch.Qos(
        config.QoS.PrefetchCount,
        config.QoS.PrefetchSize,
        config.QoS.Global,
    )
    if err != nil {
        log.Fatalf("Failed to set QoS: %v", err)
    }
    
    // Declare queue
    q, err := ch.QueueDeclare(
        config.Queue.Name,
        config.Queue.Durable,
        config.Queue.AutoDelete,
        config.Queue.Exclusive,
        config.Queue.NoWait,
        config.Queue.Arguments,
    )
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }
    
    // Create worker pool
    workerPool := make(chan struct{}, config.RPC.MaxWorkers)
    for i := 0; i < config.RPC.MaxWorkers; i++ {
        workerPool <- struct{}{}
    }
    
    // Consume messages
    msgs, err := ch.Consume(
        q.Name,
        config.Consumer.Tag,
        config.Consumer.AutoAck,
        config.Consumer.Exclusive,
        config.Consumer.NoLocal,
        config.Consumer.NoWait,
        config.Consumer.Args,
    )
    if err != nil {
        log.Fatalf("Failed to register consumer: %v", err)
    }
    
    log.Println("RPC Server started. Waiting for requests...")
    
    // Process messages
    go func() {
        for msg := range msgs {
            <-workerPool // Acquire worker
            
            go func(d amqp091.Delivery) {
                defer func() {
                    workerPool <- struct{}{} // Release worker
                }()
                
                log.Printf("Received: %s", d.Body)
                
                // Process with timeout
                responseCh := make(chan []byte, 1)
                
                go func() {
                    // Simulate processing
                    response := []byte("Processed: " + string(d.Body))
                    responseCh <- response
                }()
                
                // Wait for response or timeout
                select {
                case response := <-responseCh:
                    // Send response
                    err = ch.Publish(
                        "",
                        d.ReplyTo,
                        false,
                        false,
                        amqp091.Publishing{
                            ContentType:   "text/plain",
                            CorrelationId: d.CorrelationId,
                            Body:          response,
                        },
                    )
                    if err != nil {
                        log.Printf("Failed to send response: %v", err)
                    }
                    
                    d.Ack(false)
                    
                case <-time.After(config.RPC.ProcessTimeout):
                    log.Printf("Request timeout")
                    d.Nack(false, false)
                }
            }(msg)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down...")
}
