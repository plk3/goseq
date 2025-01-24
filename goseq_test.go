package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestRunParallelProcessing_Success(t *testing.T) {
	getData := func() (<-chan int, <-chan error) {
		dataChan := make(chan int)
		errChan := make(chan error)
		go func() {
			defer close(dataChan)
			defer close(errChan)
			for i := 1; i <= 10; i++ {
				dataChan <- i
			}
		}()
		return dataChan, errChan
	}

	processData := func(x int) (string, error) {
		return fmt.Sprintf("value: %d", x*2), nil
	}

	results, err := ProcessParallelOrdered(getData, processData, 3)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expected := []string{
		"value: 2", "value: 4", "value: 6", "value: 8", "value: 10",
		"value: 12", "value: 14", "value: 16", "value: 18", "value: 20",
	}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, but got %v", expected, results)
	}
}

func TestRunParallelProcessing_ProcessError(t *testing.T) {
	getData := func() (<-chan int, <-chan error) {
		dataChan := make(chan int)
		errChan := make(chan error)
		go func() {
			defer close(dataChan)
			defer close(errChan)
			for i := 1; i <= 5; i++ {
				dataChan <- i
			}
		}()
		return dataChan, errChan
	}

	processData := func(x int) (string, error) {
		if x == 3 {
			return "", errors.New("error processing data 3")
		}
		return fmt.Sprintf("value: %d", x*2), nil
	}

	_, err := ProcessParallelOrdered(getData, processData, 3)
	if err == nil {
		t.Fatalf("Expected an error, but got nil")
	}
	if err.Error() != "error processing data 3" {
		t.Errorf("Expected specific error, but got %v", err)
	}
}

func TestRunParallelProcessing_DataError(t *testing.T) {
	getData := func() (<-chan int, <-chan error) {
		errChan := make(chan error, 1)
		go func() {
			errChan <- errors.New("data retrieval error")
			close(errChan)
		}()
		return nil, errChan
	}

	processData := func(x int) (string, error) {
		return fmt.Sprintf("value: %d", x), nil
	}

	_, err := ProcessParallelOrdered(getData, processData, 3)
	if err == nil {
		t.Fatalf("Expected an error, but got nil")
	}
	if err.Error() != "data retrieval error" {
		t.Errorf("Expected specific error, but got %v", err)
	}
}

func TestRunParallelProcessing_EmptyData(t *testing.T) {
	getData := func() (<-chan int, <-chan error) {
		dataChan := make(chan int)
		errChan := make(chan error)
		close(dataChan)
		close(errChan)
		return dataChan, errChan
	}

	processData := func(x int) (string, error) {
		return fmt.Sprintf("value: %d", x*2), nil
	}

	results, err := ProcessParallelOrdered(getData, processData, 3)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected empty result, but got: %v", results)
	}
}

func TestRunParallelProcessing_ZeroWorker(t *testing.T) {
	getData := func() (<-chan int, <-chan error) {
		dataChan := make(chan int)
		errChan := make(chan error)
		go func() {
			defer close(dataChan)
			defer close(errChan)
			for i := 1; i <= 5; i++ {
				dataChan <- i
			}
		}()
		return dataChan, errChan
	}
	processData := func(x int) (string, error) {
		return fmt.Sprintf("value: %d", x*2), nil
	}

	results, err := ProcessParallelOrdered(getData, processData, 0)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	expected := []string{}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, but got %v", expected, results)
	}
}
