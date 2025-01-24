package main

import (
	"sync"
)

type Processer[T any, R any] func(T) (R, error)

type Getter[T any] func() (<-chan T, <-chan error)

type IndexData[T any] struct {
	Index int
	Data  T
}

func ProcessParallelOrdered[T any, R any](getter Getter[T], processor Processer[T, R], numWorkers int) ([]R, error) {
	dataChan := make(chan IndexData[T])
	resultChan := make(chan IndexData[R], numWorkers)
	errChan := make(chan error, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup

	go fetchIndexedData(getter, dataChan, errChan)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			processIndexedData(processor, dataChan, resultChan, errChan)
			wg.Done()
		}()
	}
	collectedData := make([]R, 0)
	go collectIndexedResults(resultChan, done, &collectedData)
	wg.Wait()
	close(resultChan)
	<-done

	select {
	case err := <-errChan:
		return nil, err
	default:
		return collectedData, nil
	}
}

func fetchIndexedData[T any](getter Getter[T], dataChan chan<- IndexData[T], errChan chan<- error) {
	defer close(dataChan)
	dataStream, errorStream := getter()
	index := 0
	for {
		select {
		case data, ok := <-dataStream:
			if !ok {
				return
			}
			dataChan <- IndexData[T]{Index: index, Data: data}
			index++
		case err, ok := <-errorStream:
			if ok {
				select {
				case errChan <- err:
				default:
				}
			}
			return
		}
	}
}

func processIndexedData[T any, R any](processor Processer[T, R], dataChan <-chan IndexData[T], resultChan chan<- IndexData[R], errChan chan<- error) {
	for chunk := range dataChan {
		processedData, err := processor(chunk.Data)
		if err != nil {
			select {
			case errChan <- err:
			default:
			}
			return
		}
		resultChan <- IndexData[R]{Index: chunk.Index, Data: processedData}
	}
}

func collectIndexedResults[R any](resultChan <-chan IndexData[R], done chan<- struct{}, collectData *[]R) {
	defer close(done)
	for chunk := range resultChan {
		if chunk.Index >= len(*collectData) {
			for i := len(*collectData); i <= chunk.Index; i++ {
				*collectData = append(*collectData, *new(R))
			}
		}
		(*collectData)[chunk.Index] = chunk.Data
	}
}
