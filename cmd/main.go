package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	CountWorkers    int
	SizeChanPacket  int
	TickerPublisher time.Duration
	TickerMessage   time.Duration
}

const (
	DefaultCountWorkers    = 2
	DefaultSizeChanPacket  = 4
	DefaultTickerPublisher = time.Duration(500 * time.Millisecond)
	DefaultTickerMessage   = time.Duration(1 * time.Second)
)

func parseConfig() Config {
	config := Config{}

	countWorkersStr := os.Getenv("COUNT_WORKERS")
	countWorker, err := strconv.Atoi(countWorkersStr)
	if err != nil || countWorker == 0 {
		log.Println("Переменная COUNT_WORKERS задана не верно")
		countWorker = DefaultCountWorkers
	}
	config.CountWorkers = countWorker

	tickerPublisherStr := os.Getenv("TICKER_PUBLISHER")
	tickerPublisher, err := strconv.ParseInt(tickerPublisherStr, 10, 64)

	config.TickerPublisher = time.Duration(tickerPublisher) * time.Millisecond
	if err != nil || tickerPublisher == 0 {
		log.Println("Переменная TICKER_PUBLISHER задана не верно")
		config.TickerPublisher = DefaultTickerPublisher
	}

	tickerMessageStr := os.Getenv("TICKER_MESSAGE")
	tickerMessage, err := strconv.ParseInt(tickerMessageStr, 10, 64)

	config.TickerMessage = time.Duration(tickerMessage) * time.Millisecond
	if err != nil || tickerMessage == 0 {
		log.Println("Переменная TICKER_MESSAGE задана не верно")
		config.TickerMessage = DefaultTickerMessage
	}

	return config
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Конфигурируем наше приложение
	// если env не заданы используем константы
	config := parseConfig()

	chPacket := make(chan []int, DefaultSizeChanPacket)
	chBattery := make(chan [3]int, DefaultSizeChanPacket)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go WorkerBattery(ctx, &chBattery, config.TickerMessage, &wg)
	go WorkerPublisher(ctx, &chPacket, config.TickerPublisher, &wg)

	for i := 0; i < config.CountWorkers; i++ {
		wg.Add(1)
		go WorkerConsumer(ctx, &chPacket, &chBattery, &wg)
	}

	wg.Wait()

}

func WorkerConsumer(ctx context.Context, chPackets *chan []int, chResult *chan [3]int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("WorkerConsumer завершил работу")
			return
		case packet := <-*(chPackets):
			sort.Slice(packet, func(i int, j int) bool {
				return packet[i] > packet[j]
			})

			var result [3]int
			for i := 0; i < 3; i++ {
				result[i] = packet[i]
			}

			*(chResult) <- result
		}
	}
}

func WorkerBattery(ctx context.Context, ch *chan [3]int, t time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	var sum int
	var ticker *time.Ticker = time.NewTicker(t)

	for {
		select {
		case <-ctx.Done():
			log.Println("WorkerBattery завершил работу")
			return
		case res := <-*(ch):
			for _, v := range res {
				sum += v
			}

		case <-ticker.C:
			fmt.Println(sum)
		}
	}
}

func WorkerPublisher(ctx context.Context, ch *chan []int, t time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	var ticker *time.Ticker = time.NewTicker(t)

	for {
		select {
		case <-ctx.Done():
			log.Println("WorkerPublisher завершил работу")
			return
		case <-ticker.C:
			packet := generateRandomSlice(10)
			*(ch) <- packet
		}
	}
}

func generateRandomSlice(size int) []int {
	slice := make([]int, size, size)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		slice[i] = rand.Intn(10)
	}

	return slice
}
