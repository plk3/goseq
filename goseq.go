package main

import (
	"iter"
	"sync"
)

type Processer[T any, R any] func(T) (R, error)

type IndexData[T any] struct {
	Index int
	Data  T
}

func ProcessParallelOrdered[T any, R any](getter iter.Seq[T], processor Processer[T, R], numWorkers int) ([]R, error) {
	dataChan := make(chan IndexData[T], numWorkers)
	resultChan := make(chan IndexData[R], numWorkers)
	errChan := make(chan error, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup

	go func() {
		defer close(dataChan)
		index := 0
		for data := range getter {
			dataChan <- IndexData[T]{Index: index, Data: data}
			index++
		}
	}()
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
			*collectData = append(*collectData, make([]R, chunk.Index-len(*collectData)+1)...)
		}
		(*collectData)[chunk.Index] = chunk.Data
	}
}
