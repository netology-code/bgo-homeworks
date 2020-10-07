package main

import (
	"fmt"
	"sync"
)

func merge(channels []<-chan int64) <-chan int64 {
	wg := sync.WaitGroup{}
	wg.Add(len(channels))

	result := make(chan int64)

	// проходимся по всем каналам
	for _, ch := range channels {
		// для каждого канала запускаем горутину, которая будет читать данные из этого канала
		go func(ch <-chan int64) {
			// когда канал ch закроется, мы выйдем из цикла for
			for value := range ch {
				result <- value
			}
			wg.Done()
		}(ch)
	}
	// отдельная горутина, которая ждёт, пока все каналы
	// из которых мы читаем закроются (через WaitGroup)
	go func() {
		wg.Wait()
		// после чего сама закрывает канал в который пишет
		close(result)
	}()
	// возвращаем итоговый канал
	return result
}

func main() {
	count := 9_999_999 // всего транзакций
	transactions := make([]int64, count)
	for i := range transactions {
		transactions[i] = 1_00 // каждая транзакция 1 рубль
	}

	parts := 10
	partSize := len(transactions) / parts

	// слайс каналов - каждой горутине дадим по каналу, она в него будет писать
	channels := make([]<-chan int64, parts)

	for i := 0; i < parts; i++ {
		var part []int64

		if i != parts-1 {
			part = transactions[i*partSize : (i+1)*partSize]
		} else {
			part = transactions[i*partSize:]
		}

		ch := make(chan int64)
		channels[i] = ch

		go func(data []int64, ch chan<- int64) {
			sum := int64(0)
			for _, datum := range data {
				sum += datum
			}
			// горутина пишет одно значение в канал
			// на самом деле не обязательно одно: вы можете поместить в цикл, чтобы осуществлять только фильтрацию
			ch <- sum
			// и закрывает его
			close(ch)
		}(part, ch)
	}

	total := int64(0)
	// "собираем" все каналы в один и из него читаем итоговый результат
	// когда канал закроется, то выйдем из for
	for value := range merge(channels) {
		total += value
	}

	fmt.Println(total)
}
