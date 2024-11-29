package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: pml -input=input.properties -output=output.yml")
		flag.PrintDefaults()
	}

	inputFile := flag.String("input", "", "Path to the input .properties file (required)")
	outputFile := flag.String("output", "", "Path to the output .yml file (required)")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Error: Both -input and -output flags must be specified.")
		flag.Usage()
		os.Exit(1)
	}

	properties, err := readProperties(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", *inputFile, err)
		os.Exit(1)
	}

	nestedMap := createNestedMap(properties)

	if err := writeYAMLWithSpaces(nestedMap, *outputFile); err != nil {
		fmt.Printf("Error writing YAML file %s: %v\n", *outputFile, err)
		os.Exit(1)
	}

	fmt.Printf("Conversion complete: %s -> %s\n", *inputFile, *outputFile)
}

func readProperties(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	properties := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		properties[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return properties, nil
}

func createNestedMap(flatMap map[string]string) map[string]interface{} {
	nestedMap := make(map[string]interface{})

	for key, value := range flatMap {
		parts := strings.Split(key, ".")
		currentMap := nestedMap

		for i, part := range parts {
			if i == len(parts)-1 {
				currentMap[part] = value
			} else {
				if _, exists := currentMap[part]; !exists {
					currentMap[part] = make(map[string]interface{})
				}
				currentMap = currentMap[part].(map[string]interface{})
			}
		}
	}

	return nestedMap
}

func writeYAMLWithSpaces(data map[string]interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	encoder := yaml.NewEncoder(file)
	defer func(encoder *yaml.Encoder) {
		err := encoder.Close()
		if err != nil {
			panic(err)
		}
	}(encoder)

	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}
