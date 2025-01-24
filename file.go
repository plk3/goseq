package main

import (
	"bufio"
	"io"
	"os"
)

func ProcessParallelLines(reader io.Reader, processor Processer[string, string], numWorkers int) ([]string, error) {
	getter := func() (<-chan string, <-chan error) {
		out := make(chan string)
		go func() {
			defer close(out)
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				out <- scanner.Text()
			}
		}()
		return out, nil
	}
	return ProcessParallelOrdered(getter, processor, numWorkers)
}

func ProcessParallelLinesInChunks(reader io.Reader, processor Processer[[]string, []string], chunkSize int, numWorkers int) ([]string, error) {
	getter := func() (<-chan []string, <-chan error) {
		out := make(chan []string)
		go func() {
			defer close(out)
			scanner := bufio.NewScanner(reader)
			lines := make([]string, 0, chunkSize)
			lineCount := 0
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
				lineCount++
				if lineCount == chunkSize {
					out <- lines
					lines = make([]string, 0, chunkSize)
					lineCount = 0
				}
			}
			if len(lines) > 0 {
				out <- lines
			}
		}()
		return out, nil
	}
	data, err := ProcessParallelOrdered(getter, processor, numWorkers)
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
	getter := func() (<-chan []byte, <-chan error) {
		out := make(chan []byte, len(filePaths))
		errChan := make(chan error, 1)
		go func() {
			defer close(out)
			defer close(errChan)
			for _, filePath := range filePaths {
				file, err := os.Open(filePath)
				if err != nil {
					errChan <- err
					return
				}
				defer file.Close()
				data, err := io.ReadAll(file)
				if err != nil {
					errChan <- err
					return
				}
				out <- data
			}

		}()
		return out, errChan
	}

	data, err := ProcessParallelOrdered(getter, processor, numWorkers)
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
