// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
)

const resultFilePath = "trained_logistic_regression_model.joblib"

var (
	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")
)

type Metadata map[string]interface{}

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	Run(ctx context.Context, cmp Computation) (string, error)
	Algo(ctx context.Context, algorithm []byte) (string, error)
	Data(ctx context.Context, dataset []byte) (string, error)
	Result(ctx context.Context) ([]byte, error)
}

type agentService struct {
	computation Computation
	algorithms  [][]byte
	datasets    [][]byte
	result      []byte
}

var _ Service = (*agentService)(nil)

// New instantiates the agent service implementation.
func New() Service {
	return &agentService{}
}

func (as *agentService) Run(ctx context.Context, cmp Computation) (string, error) {
	cmpJSON, err := json.Marshal(cmp)
	if err != nil {
		return "", err
	}

	as.computation = cmp

	return string(cmpJSON), nil // return the JSON string as the function's string return value
}

func (as *agentService) Algo(ctx context.Context, algorithm []byte) (string, error) {
	// Implement the logic for the Algo method based on your requirements
	// Use the provided ctx and algorithm parameters as needed

	as.algorithms = append(as.algorithms, algorithm)

	// Perform some processing on the algorithm byte array
	// For example, generate a unique ID for the algorithm
	algorithmID := "algo123"

	// Return the algorithm ID or an error
	return algorithmID, nil
}

func (as *agentService) Data(ctx context.Context, dataset []byte) (string, error) {
	// Implement the logic for the Data method based on your requirements
	// Use the provided ctx and dataset parameters as needed

	as.datasets = append(as.datasets, dataset)

	// Perform some processing on the dataset string
	// For example, generate a unique ID for the dataset
	datasetID := "dataset456"

	// Return the dataset ID or an error
	return datasetID, nil
}

const marker = "===MODEL_MARKER==="

func (as *agentService) Result(ctx context.Context) ([]byte, error) {
	// Implement the logic for the Result method based on your requirements
	// Use the provided ctx parameter as needed

	// Perform some processing to retrieve the computation result file
	// For example, read the file from storage or generate a dummy result
	result, err := run(as.algorithms[0], as.datasets[0], resultFilePath)
	if err != nil {
		return nil, fmt.Errorf("error performing computation: %v", err)
	}
	as.result = result

	// Return the result file or an error
	return as.result, nil
}

func run(algoContent []byte, dataContent []byte, resultPath string) ([]byte, error) {
	// Construct the Python script content with CSV data as a command-line argument
	script := string(algoContent)
	data := string(dataContent)

	// Run the Python script with the script and data as input
	cmd := exec.Command("python3", "-c", script, data, resultPath)

	// Capture the command's standard output and error
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting Python script: %v", err)
	}

	// Create memory-based output and error writers
	var stdoutBuffer, stderrBuffer bytes.Buffer

	// Read and print the standard output and error
	go func() {
		_, _ = io.Copy(&stdoutBuffer, stdout)
	}()
	go func() {
		_, _ = io.Copy(&stderrBuffer, stderr)
	}()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("python script execution error: %v", err)
	}

	fmt.Println("Python script execution completed.")

	// Find the markers and extract the model bytes
	resultBuffer, err := io.ReadAll(&stdoutBuffer)
	if err != nil {
		log.Fatalf("Error reading stdout: %v", err)
	}

	startMarker := []byte(marker)
	endMarker := []byte(marker)

	result, err := extractResult(resultBuffer, startMarker, endMarker)
	if err != nil {
		log.Println(err)
	} else {
		// Now you can use modelBytes in your Go program
		// For example, you can deserialize the model using joblib
	}

	return result, nil
}

func extractResult(output []byte, startMarker, endMarker []byte) ([]byte, error) {
	start := bytes.Index(output, startMarker)
	end := bytes.LastIndex(output, endMarker)

	if start != -1 && end != -1 && start < end {
		modelBytes := output[start+len(startMarker) : end]
		return modelBytes, nil
	}

	return nil, errors.New("model marker not found in output")
}
