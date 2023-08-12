package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"
)

type Config struct {
	CountWorkers    int
	SizeChanPacket  int
	TickerPublisher time.Duration
	TickerMessage   time.Duration
}

const (
	DefaultCountWorkers = 10
)

func main() {
	config := Config{}

	countWorkersStr := os.Getenv("COUNT_WORKERS")
	countWorker, err := strconv.Atoi(countWorkersStr)
	if err != nil || countWorker == 0 {
		log.Println("Переменная COUNT_WORKERS задана не верно")
		countWorker = DefaultCountWorkers
	}

	config.CountWorkers = countWorker

	chPacket := make(chan []int, config.SizeChanPacket)
	// Делаем второй канал такого же размера чтобы небыло блокировки
	chBattery := make(chan [3]int, config.SizeChanPacket)

	//w := sync.WaitGroup{}

	ctx := context.Background()
	// TODO пробросить контекст попозже нормальный
	go WorkerBattery(ctx, &chBattery, config.TickerMessage)
	go WorkerPublisher(ctx, &chPacket, config.TickerPublisher)

	for i := 0; i < config.CountWorkers; i++ {
		//w.Add(1)
		go WorkerConsumer(ctx, &chPacket, &chBattery)
	}

	time.Sleep(time.Second * 300)
}

func WorkerConsumer(ctx context.Context, chPackets *chan []int, chResult *chan [3]int) {
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

func WorkerBattery(ctx context.Context, ch *chan [3]int, t time.Duration) {
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

func WorkerPublisher(ctx context.Context, ch *chan []int, t time.Duration) {
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
