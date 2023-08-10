package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pbtools [path]")
		return
	}

	dirPath := os.Args[1]

	protoFiles, err := getAllProtoFiles(dirPath)
	if err != nil {
		fmt.Println("Error getting .proto files:", err)
		return
	}

	for _, file := range protoFiles {
		processFile(file)
	}
}

func getAllProtoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".proto" {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func processFile(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", filePath, err)
		return
	}

	updatedContent := string(content)
	messageRegex := regexp.MustCompile(`message\s+(\w+)\s+\{([^\}]+)\}`)
	fieldRegex := regexp.MustCompile(`(\s*[\w]+\s+[\w]+\s*=\s*)(\d+)`)

	matches := messageRegex.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		messageContent := match[2]
		fieldMatches := fieldRegex.FindAllStringSubmatchIndex(messageContent, -1)

		var indices []int
		for _, fieldMatch := range fieldMatches {
			start, end := fieldMatch[4], fieldMatch[5]
			index, _ := strconv.Atoi(messageContent[start:end])
			indices = append(indices, index)
		}

		mapping := createMapping(indices)
		messageReplacement := messageContent
		for oldIdx, newIdx := range mapping {
			replaceStr := fmt.Sprintf("${1}%d", newIdx)
			messageReplacement = fieldRegex.ReplaceAllStringFunc(messageReplacement, func(s string) string {
				fieldIdx, _ := strconv.Atoi(fieldRegex.ReplaceAllString(s, "$2"))
				if fieldIdx == oldIdx {
					return fieldRegex.ReplaceAllString(s, replaceStr)
				}
				return s
			})
		}
		updatedContent = strings.Replace(updatedContent, messageContent, messageReplacement, 1)
	}

	if updatedContent != string(content) {
		err = os.WriteFile(filePath, []byte(updatedContent), 0644)
		if err != nil {
			fmt.Println("Error writing file:", filePath, err)
		}
	}
}

func createMapping(indices []int) map[int]int {
	sort.Ints(indices)

	mapping := make(map[int]int)
	missing := 1
	for _, index := range indices {
		if index != missing {
			mapping[index] = missing
		}
		missing++
	}

	return mapping
}
