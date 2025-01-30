package main

import (
	"bufio"
	"io"
	"os"

	"log"
)

func ProcessParallelLines(reader io.Reader, processor Processer[string, string], numWorkers int) ([]string, error) {
	getIter := func(yield func(string) bool) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				break
			}
		}
	}
	return ProcessParallelOrdered(getIter, processor, numWorkers)
}

func ProcessParallelLinesInChunks(reader io.Reader, processor Processer[[]string, []string], chunkSize int, numWorkers int) ([]string, error) {
	getIter := func(yield func([]string) bool) {
		chunk := make([]string, 0, chunkSize)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			chunk = append(chunk, scanner.Text())

			if len(chunk) == chunkSize {
				if !yield(chunk) {
					return
				}
				chunk = make([]string, 0, chunkSize)
			}
		}

		if len(chunk) > 0 {
			yield(chunk)
		}
	}
	data, err := ProcessParallelOrdered(getIter, processor, numWorkers)
	if err != nil {
		return nil, err
	}
	totalSize := 0
	for _, inner := range data {
		totalSize += len(inner)
	}
	result := make([]string, 0, totalSize)
	for _, inner := range data {
		result = append(result, inner...)
	}
	return result, nil
}

func ProcessParallelFiles(filePaths []string, processor Processer[[]byte, []byte], numWorkers int) ([]byte, error) {
	getIter := func(yield func([]byte) bool) {
		for _, path := range filePaths {
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("read error: %v\n", err)
				continue
			}
			if !yield(data) {
				return
			}
		}
	}

	data, err := ProcessParallelOrdered(getIter, processor, numWorkers)
	if err != nil {
		return nil, err
	}
	totalSize := 0
	for _, inner := range data {
		totalSize += len(inner)
	}
	result := make([]byte, 0, totalSize)
	for _, inner := range data {
		result = append(result, inner...)
	}
	return result, nil
}
