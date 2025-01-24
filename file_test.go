package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func sampleProcessor(lines []string) ([]string, error) {
	var result []string
	for _, line := range lines {
		if strings.HasSuffix(line, "2") {
			continue
		}
		result = append(result, strings.ToUpper(line))
	}
	return result, nil
}

func TestProcessParallelLinesInChunks(t *testing.T) {
	input := ""
	for i := 0; i < 1000; i++ {
		input += "line" + strconv.Itoa(i) + "\n"
	}
	reader := strings.NewReader(input)

	expectedOutput := []string{}
	for i := 0; i < 1000; i++ {
		if i%10 == 2 {
			continue
		}
		expectedOutput = append(expectedOutput, "LINE"+strconv.Itoa(i))
	}
	output, err := ProcessParallelLinesInChunks(reader, sampleProcessor, 100, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output) != len(expectedOutput) {
		t.Fatalf("expected %d lines, got %d lines", len(expectedOutput), len(output))
	}
	for i, line := range output {
		if line != expectedOutput[i] {
			t.Errorf("expected line %d to be %q, got %q", i, expectedOutput[i], line)
		}
	}
}

func TestProcessParallelLines(t *testing.T) {
	input := ""
	for i := 0; i < 1000; i++ {
		input += "line" + strconv.Itoa(i) + "\n"
	}
	reader := strings.NewReader(input)

	processor := func(line string) (string, error) {
		return strings.ToUpper(line), nil
	}

	expectedOutput := []string{}
	for i := 0; i < 1000; i++ {
		expectedOutput = append(expectedOutput, "LINE"+strconv.Itoa(i))
	}
	output, err := ProcessParallelLines(reader, processor, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output) != len(expectedOutput) {
		t.Fatalf("expected %d lines, got %d lines", len(expectedOutput), len(output))
	}
	for i, line := range output {
		if line != expectedOutput[i] {
			t.Errorf("expected line %d to be %q, got %q", i, expectedOutput[i], line)
		}
	}
}

func TestProcessParallelFiles(t *testing.T) {
	processor := func(input []byte) ([]byte, error) {
		return append([]byte("processed: "), input...), nil
	}
	processWithError := func(input []byte) ([]byte, error) {
		return nil, errors.New("processing error")
	}

	t.Run("Normal case: process multiple files", func(t *testing.T) {
		files, cleanup := createTempFiles(t, [][]byte{
			[]byte("file1 content"),
			[]byte("file2 content"),
			[]byte("file3 content"),
		})
		defer cleanup()

		result, err := ProcessParallelFiles(files, processor, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []byte("processed: file1 contentprocessed: file2 contentprocessed: file3 content")
		if string(result) != string(expected) {
			t.Errorf("result does not match.\nexpected: %q\nactual: %q", expected, result)
		}
	})

	t.Run("Error case: file read error occurs", func(t *testing.T) {
		files, cleanup := createTempFiles(t, [][]byte{
			[]byte("file1 content"),
		})
		defer cleanup()

		files = append(files, "invalid_file_path")

		_, err := ProcessParallelFiles(files, processor, 5)
		if err == nil {
			t.Fatal("error did not occur.")
		}
	})

	t.Run("Error case: Processor throws an error", func(t *testing.T) {
		files, cleanup := createTempFiles(t, [][]byte{
			[]byte("file1 content"),
			[]byte("file2 content"),
			[]byte("file3 content"),
		})
		defer cleanup()

		result, err := ProcessParallelFiles(files, processWithError, 5)
		if err == nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("result is not nil: %v", result)
		}
	})
}

func createTempFiles(t *testing.T, contents [][]byte) ([]string, func()) {
	var files []string
	cleanup := func() {
		for _, file := range files {
			os.Remove(file)
		}
	}
	for i, content := range contents {
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("testfile_%d", i))
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer tmpFile.Close()
		_, err = tmpFile.Write(content)
		if err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		files = append(files, tmpFile.Name())
	}
	return files, cleanup
}
