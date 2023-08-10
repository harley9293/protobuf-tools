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
	fieldRegex := regexp.MustCompile(`(\s*(?:\w+|map<\w+\s*,\s*\w+>)\s+\w+\s*=\s*)(\d+)`)

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

		ids := createReplaceIDs(indices)
		messageReplacement := messageContent
		for _, replaceID := range ids {
			oldIdx, newIdx := replaceID[0], replaceID[1]
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

func createReplaceIDs(indices []int) [][2]int {
	sort.Ints(indices)

	ids := make([][2]int, len(indices))
	missing := 1
	for _, index := range indices {
		if index != missing {
			ids = append(ids, [2]int{index, missing})
		}
		missing++
	}

	return ids
}
