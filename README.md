# goseq

`goseq` is a utility package for implementing parallel processing in Go **while preserving the original order** of the input data stream. This package provides the functionality to execute processing in parallel with a specified number of workers, and returns the results **in the same sequence as the input**.

## Features

- **Parallel Processing:** Launches multiple workers to process a data stream in parallel.
- **Order Guarantee:** The processing results are returned while **strictly preserving the original order** of the input data stream.
- **Error Handling:** Returns the first error that occurs during processing.
- **Genericity:** Can accept data of any type `T` and return results of any type `R`.

## Install

Install the package with:

```bash
go get github.com/plk3/goseq
```

Import it with:

```go
import "github.com/plk3/goseq"
```

### Example

```go
package main

import (
	"errors"
	"fmt"
    "github.com/plk3/goseq" 
)

func exampleGetter() (<-chan int, <-chan error) {
	dataStream := make(chan int)
	errorStream := make(chan error)

	go func() {
		defer close(dataStream)
		defer close(errorStream)
		for i := 0; i < 10; i++ {
			dataStream <- i
		}
	}()

	return dataStream, errorStream
}

func exampleProcessor(input int) (int, error) {
	return input * 2, nil
}

func main() {
	results, err := goseq.ProcessParallelOrdered(
		exampleGetter,
		exampleProcessor,
		4,
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Results:", results)
    // Results: [0 2 4 6 8 10 12 14 16 18]
}   
```

## Func

### `ProcessParallelOrdered[T any, R any]`
```go
func ProcessParallelOrdered[T any, R any](getter Getter[T], processor Processer[T, R], numWorkers int) ([]R, error)
```

### `ProcessPallelLines`
```go
func ProcessPallelLines(reader io.Reader, processor Processer[string, string], numWorkers int) ([]string, error) 
```

### `ProcessParallelLinesInChunks`
```go
func ProcessParallelLinesInChunks(reader io.Reader, processor Processer[[]string, []string], chunkSize int, numWorkers int) ([]string, error) 
```

### `ProcessParallelFiles`
```go
func ProcessParallelFiles(filePaths []string, processor Processer[[]byte, []byte], numWorkers int) ([]byte, error) 
```


## Interfaces

### `Processer[T any, R any]`

A function type that processes data `T` and returns a result `R` and an error.

```go
type Processer[T any, R any] func(T) (R, error)
```
### `Getter[T any]`
A function type that returns a data stream (<-chan T) and an error stream (<-chan error). This function is responsible for providing the input data for parallel processing.
```go
type Getter[T any] func() (<-chan T, <-chan error)
```
