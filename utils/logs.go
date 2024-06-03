package utils

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func MonitorLogs(logPath, containerName string, usePodman *bool, dockerSocketPath string) {
	fmt.Println(Info("Monitoring logs mongodb..."))
	var cmd *exec.Cmd
	var file *os.File
	var err error
	targetPath := "mongo_logs.txt"
	if containerName != "" && logPath != "" {
		fmt.Println(Info("Detected docker container usage with custom path to log file."))
		if *usePodman {
			cmd = exec.Command("podman", "exec", containerName, "cat", logPath)
		} else {
			cmd = exec.Command("docker", "exec", containerName, "cat", logPath)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}

		file, err = os.Create(targetPath) // Set the file variable
		if err != nil {
			log.Fatalf("failed to create file: %s", err)
		}
		defer file.Close()

		_, err = file.WriteString(string(output))
		if err != nil {
			log.Fatalf("error writing to file: %s", err)
		}
	} else if containerName != "" && logPath == "" {
		fmt.Println(Info("Detected docker container usage with default stream logging."))
		// defaultDockerSocketPath := "unix:///var/run/docker.sock"

		// dockerSocketPath := defaultDockerSocketPath
		// if customDockerSocketPath != "" {
		// 	dockerSocketPath = customDockerSocketPath
		// }
		fmt.Println(dockerSocketPath)
		// fmt.Println(customDockerSocketPath)
		cli, err := client.NewClientWithOpts(
			client.WithHost(dockerSocketPath),
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			log.Fatalf("Failed to create containerization client: %s", err)
		}
		options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Timestamps: false}
		out, err := cli.ContainerLogs(context.Background(), containerName, options)
		if err != nil {
			log.Fatalf("Failed to get container logs: %s", err)
		}
		defer out.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, out); err != nil {
			log.Fatalf("Failed to read container logs: %s", err)
		}

		filteredData := FilterPrintableASCII(buf.Bytes())
		file, err = os.Create(targetPath) // Set the file variable
		if err != nil {
			log.Fatalf("failed to create file: %s", err)
		}
		defer file.Close()

		_, err = file.WriteString(string(filteredData))
		if err != nil {
			log.Fatalf("error writing to file: %s", err)
		}

	} else {
		fmt.Println(Info("Detected default way with local mongodb log file."))
		content, err := os.ReadFile(logPath)
		if err != nil {
			log.Fatalf("failed to read file: %s", err)
		}

		err = os.WriteFile(targetPath, content, 0644)
		if err != nil {
			log.Fatalf("failed to write to file: %s", err)
		}
	}

	var collScanLines []string

	content, err := os.ReadFile(targetPath) // Use the opened file for reading
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}

	logs := string(content)

	// Split the logs by newlines
	logLines := strings.Split(logs, "\n")

	for _, line := range logLines {
		if strings.Contains(line, "COLLSCAN") {
			collScanLines = append(collScanLines, line)
		}
	}

	outputFilePath := "colout.txt"
	fmt.Println(Info("Output file path: ", outputFilePath))

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("failed to create output file: %s", err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range collScanLines {
		fmt.Fprintln(writer, line)
	}
	if err := writer.Flush(); err != nil {
		log.Fatalf("error while writing to file: %s", err)
	}
	err = os.Remove(targetPath)
	if err != nil {
		log.Fatalf("failed to remove file: %s", err)
	}
	fmt.Println(Info("--- Collscan Lines Written to ", outputFilePath, "---"))
}
