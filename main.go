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
		fmt.Println("Usage: pml -i=input.file -o=output.file -f=format")
		flag.PrintDefaults()
	}

	inputFile := flag.String("i", "", "Path to the input file (required)")
	outputFile := flag.String("o", "", "Path to the output file (required)")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Error: -i and -o flags must be specified.")
		flag.Usage()
		os.Exit(1)
	}

	// Determine file extension and proceed accordingly
	if strings.HasSuffix(*inputFile, ".properties") {
		properties, err := readProperties(*inputFile)
		if err != nil {
			fmt.Printf("Error reading properties file %s: %v\n", *inputFile, err)
			os.Exit(1)
		}

		nestedMap := createNestedMap(properties)

		if err := writeYAMLWithSpaces(nestedMap, *outputFile); err != nil {
			fmt.Printf("Error writing YAML file %s: %v\n", *outputFile, err)
			os.Exit(1)
		}

		fmt.Printf("Conversion complete: %s -> %s\n", *inputFile, *outputFile)

	} else if strings.HasSuffix(*inputFile, ".yml") || strings.HasSuffix(*inputFile, ".yaml") {
		nestedMap, err := readYAML(*inputFile)
		if err != nil {
			fmt.Printf("Error reading YAML file %s: %v\n", *inputFile, err)
			os.Exit(1)
		}

		properties := flattenYAML(nestedMap)

		if err := writeProperties(properties, *outputFile); err != nil {
			fmt.Printf("Error writing .properties file %s: %v\n", *outputFile, err)
			os.Exit(1)
		}

		fmt.Printf("Conversion complete: %s -> %s\n", *inputFile, *outputFile)

	} else {
		fmt.Println("Error: Invalid file type. Please provide a .properties or .yml/.yaml file.")
		flag.Usage()
		os.Exit(1)
	}
}

func readProperties(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func readYAML(filename string) (map[string]interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func flattenYAML(nestedMap map[string]interface{}) map[string]string {
	properties := make(map[string]string)
	flattenMap("", nestedMap, properties)
	return properties
}

func flattenMap(prefix string, m map[string]interface{}, properties map[string]string) {
	for key, value := range m {
		newKey := prefix + key
		if subMap, ok := value.(map[string]interface{}); ok {
			flattenMap(newKey+".", subMap, properties)
		} else {
			properties[newKey] = fmt.Sprintf("%v", value)
		}
	}
}

func writeProperties(properties map[string]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for key, value := range properties {
		line := fmt.Sprintf("%s=%s\n", key, value)
		if _, err := writer.WriteString(line); err != nil {
			return err
		}
	}

	return writer.Flush()
}
