package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	broker = "localhost:29092"
	version  = "2.1.1"
	group    = "lab-cgr-payment-service-0"
	clientID = "lab-0"
	topic    = "lab-payment-requested-n"
)

func daemon(metricCH, consumerCH chan struct{}) {
	interruptionCH := make(chan os.Signal, 1)
	signal.Notify(interruptionCH, syscall.SIGINT, syscall.SIGTERM)
	select {
	case t := <-interruptionCH:
		log.Printf("caught signal: %s\n", t.String())
		consumerCH <- struct{}{}
		metricCH <- struct{}{}
	}
}

func metric(terminationCH chan struct{}) {
	log.Println("started metric server")

	routeHandler := http.NewServeMux()
	routeHandler.Handle("/metrics", promhttp.Handler())
	metricSrv := http.Server{
		Addr:    ":9200",
		Handler: routeHandler,
	}
	metricSrv.RegisterOnShutdown(func() {
		log.Println("shutting down metric server...")
	})

	go func() {
		err := metricSrv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	<-terminationCH
	err := metricSrv.Shutdown(context.Background())
	if err != nil {
		log.Println("error while shutting down metric server", err)
	}
}

func main() {
	metricCH := make(chan struct{})
	consumerCH := make(chan struct{})

	go daemon(metricCH, consumerCH)

	go metric(metricCH)
	startConsumer(consumerCH)
}

func startConsumer(terminationCH chan struct{}) {
	log.Println("started payment service...")
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg := sarama.NewConfig()
	kafkaVersion, _ := sarama.ParseKafkaVersion(version)
	cfg.Version = kafkaVersion
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.ClientID = clientID

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup([]string{broker}, group, cfg)
	if err != nil {
		log.Panicf("error creating consumer group client: %v", err)
	}
	log.Printf("created a consumer group: %s, topic: %s", strings.Split(broker, ",")[0], topic)

	consumer := Consumer{
		ready:       make(chan bool),
		interrupted: make(chan bool),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
		}()
		for {
			log.Println("starting consumer...")
			err := client.Consume(ctx, []string{topic}, &consumer)
			if err != nil {
				log.Panicf("error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}

			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	log.Println("consumer completely started...")

process:
	for {
		select {
		case <-terminationCH:
			consumer.interrupted <- true
			break process

		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			break process
		}
	}

	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("error closing client: %v", err)
	}

	log.Println("consumer closed gracefully")
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready       chan bool
	interrupted chan bool
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	log.Println("consumer up and listening...")
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("cleansing consumer group...")
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// msgCH := sequentialProcessing()
	// msgCH := concurrent()
	// msgCH := bufferedConcurrent()
	// msgCH := safeConcurrent1()
	// msgCH := safeConcurrent2()
	// defer close(msgCH)

consuming:
	for {
		select {
		case <-consumer.interrupted:
			break consuming
		default:
			msg, open := <-claim.Messages()
			if !open {
				break consuming
			}

			session.MarkMessage(msg, "")
			fullPaymentLogic(msg)
			// go fullPaymentLogic(msg)
			// msgCH <- msg
		}
	}

	return nil
}

func sequentialProcessing() chan *sarama.ConsumerMessage {
	msgCH := make(chan *sarama.ConsumerMessage)
	go func() {
		for message := range msgCH {
			fullPaymentLogic(message)
		}
	}()

	return msgCH
}

func concurrent() chan *sarama.ConsumerMessage {
	msgCH := make(chan *sarama.ConsumerMessage)
	go func() {
		for message := range msgCH {
			go fullPaymentLogic(message)
		}
	}()

	return msgCH
}

func bufferedConcurrent() chan *sarama.ConsumerMessage {
	msgCH := make(chan *sarama.ConsumerMessage, 1000)
	go func() {
		for message := range msgCH {
			go fullPaymentLogic(message)
		}
	}()

	return msgCH
}

func safeConcurrent1() chan *sarama.ConsumerMessage {
	msgCH := make(chan *sarama.ConsumerMessage)
	for i := 1; i <= 2000; i++ {
		go func() {
			for message := range msgCH {
				fullPaymentLogic(message)
			}
		}()
	}

	return msgCH
}

func safeConcurrent2() chan *sarama.ConsumerMessage {
	msgCH := make(chan *sarama.ConsumerMessage, 2000)
	for i := 1; i <= 2000; i++ {
		go func() {
			for message := range msgCH {
				fullPaymentLogic(message)
			}
		}()
	}

	return msgCH
}

func safeConcurrent3() chan *sarama.ConsumerMessage {
	msgCH := make(chan *sarama.ConsumerMessage)
	for i := 1; i <= 2000; i++ {
		go func() {
			for {
				select {
				case msg := <-msgCH:
					fullPaymentLogic(msg)
				}
			}
		}()
	}

	return msgCH
}

func fullPaymentLogic(message *sarama.ConsumerMessage) {
	log.Printf("handle payment request, partition: %d offset: %d", message.Partition, message.Offset)
	time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
	log.Printf("payment request completed, partition: %d offset: %d", message.Partition, message.Offset)
}
