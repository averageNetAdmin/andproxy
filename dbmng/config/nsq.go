package config

// import (
// 	"log"
// 	"os"
// 	"os/signal"
// 	"syscall"

// 	"github.com/nsqio/go-nsq"
// )

// func ConnectNSQ() {
// 	// // Instantiate a consumer that will subscribe to the provided channel.
// 	// config := nsq.NewConfig()
// 	// consumer, err := nsq.NewConsumer("server", "create", config)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	// // Set the Handler for messages received by this Consumer. Can be called multiple times.
// 	// // See also AddConcurrentHandlers.
// 	// consumer.AddHandler(&myMessageHandler{})

// 	// // Use nsqlookupd to discover nsqd instances.
// 	// // See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
// 	// err = consumer.ConnectToNSQLookupd("localhost:4161")
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	// // wait for signal to exit
// 	// sigChan := make(chan os.Signal, 1)
// 	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
// 	// <-sigChan

// 	// // Gracefully stop the consumer.
// 	// consumer.Stop()

// 	// Instantiate a consumer that will subscribe to the provided channel.
// 	config := nsq.NewConfig()
// 	consumer, err := nsq.NewConsumer("server", "create", config)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Set the Handler for messages received by this Consumer. Can be called multiple times.
// 	// See also AddConcurrentHandlers.
// 	consumer.AddHandler(&myMessageHandler{})

// 	// Use nsqlookupd to discover nsqd instances.
// 	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
// 	err = consumer.ConnectToNSQLookupd("localhost:4161")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// wait for signal to exit
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
// 	<-sigChan

// 	// Gracefully stop the consumer.
// 	consumer.Stop()
// }
