package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var testFilePath string
var createTestFileOnce sync.Once
var testFileCreated bool

func TestMain(m *testing.M) {
	createTestFileOnce.Do(func() {
		path, err := createTestFile()
		if err != nil {
			panic("failed to create test file: " + err.Error())
		}
		testFilePath = path
		testFileCreated = true
	})

	code := m.Run()

	if testFileCreated {
		err := os.Remove(testFilePath)
		if err != nil {
			panic("failed to remove test file: " + err.Error())
		}
	}

	os.Exit(code)
}

func ProcessLines(reader io.Reader, processor Processer[[]string, []string]) ([]string, error) {
	if reader == nil {
		return nil, errors.New("reader must not be nil")
	}
	if processor == nil {
		return nil, errors.New("processor must not be nil")
	}

	scanner := bufio.NewScanner(reader)
	var result []string

	for scanner.Scan() {
		line := scanner.Text()
		processedLines, err := processor([]string{line})
		if err != nil {
			return nil, err
		}
		result = append(result, processedLines...)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func createTestFile() (string, error) {
	tmpFile, err := os.CreateTemp("", "testfile-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	fileSize := int64(3 * 1024 * 1024 * 1024)
	line := "this is a long line to simulate a realistic scenario\n"
	lineBytes := []byte(line)

	buffer := make([]byte, 0, 100*1024*1024)
	for written := int64(0); written < fileSize; {
		buffer = append(buffer, lineBytes...)
		if int64(len(buffer)) >= 100*1024*1024 {
			n, err := tmpFile.Write(buffer)
			if err != nil {
				return "", err
			}
			written += int64(n)
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		_, err := tmpFile.Write(buffer)
		if err != nil {
			return "", err
		}
	}

	return tmpFile.Name(), nil
}

func BenchmarkProcessParallelLinesInChunks(b *testing.B) {
	processor := func(lines []string) ([]string, error) {
		result := make([]string, 0, len(lines))
		for _, line := range lines {
			result = append(result, strings.ToUpper(line))
		}
		return result, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFilePath)
		if err != nil {
			b.Fatalf("failed to open test file: %v", err)
		}
		defer file.Close()
		_, err = ProcessParallelLinesInChunks(file, processor, 10000, runtime.NumCPU()*10)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkProcessParallelLine(b *testing.B) {
	processor := func(line string) (string, error) {
		return strings.ToUpper(line), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFilePath)
		if err != nil {
			b.Fatalf("failed to open test file: %v", err)
		}
		defer file.Close()
		_, err = ProcessParallelLines(file, processor, runtime.NumCPU()*10)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkProcessLines(b *testing.B) {
	processor := func(lines []string) ([]string, error) {
		result := make([]string, 0, len(lines))
		for _, line := range lines {
			result = append(result, strings.ToUpper(line))
		}
		return result, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFilePath)
		if err != nil {
			b.Fatalf("failed to open test file: %v", err)
		}
		defer file.Close()
		_, err = ProcessLines(file, processor)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
