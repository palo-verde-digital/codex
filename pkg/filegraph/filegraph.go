package filegraph

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	inKey  = "in"
	outKey = "out"

	titleKey = "titles"
	pathKey  = "paths"
)

type NodeValue []int
type EdgeValue [][]int

type NodeProperty []string
type EdgeProperty [][]string

type graph struct {
	Size  int      `json:"size"`
	Edges [][]bool `json:"edges"`

	NodeValues map[string]NodeValue `json:"nodeValues"`
	EdgeValues map[string]EdgeValue `json:"edgeValues"`

	NodeProperties map[string]NodeProperty `json:"nodeProperties"`
	EdgeProperties map[string]EdgeProperty `json:"edgeProperties"`
}

type jsonGraph struct {
	Graph []*jsonNode `json:"graph"`
}

type jsonNode struct {
	Id        string   `json:"id"`
	Neighbors []string `json:"neighbors"`
}

func initGraph(size int) *graph {
	edges := make([][]bool, size)
	for i := range edges {
		edges[i] = make([]bool, size)
	}

	return &graph{
		Size:  size,
		Edges: edges,

		NodeValues: make(map[string]NodeValue),
		EdgeValues: make(map[string]EdgeValue),

		NodeProperties: make(map[string]NodeProperty),
		EdgeProperties: make(map[string]EdgeProperty),
	}
}

func Create(sourceDir string) error {
	workDir := sourceDir + "_tmp"

	if err := prepareWorkingEnvironment(sourceDir, workDir); err != nil {
		return fmt.Errorf("unable to create graph: %s", err.Error())
	}

	numFiles, err := countFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("unable to create graph: %s", err.Error())
	}

	graph := initGraph(numFiles)

	if err := flattenWorkingEnvironment(numFiles, workDir, graph); err != nil {
		return fmt.Errorf("unable to create graph: %s", err.Error())
	}

	nodes := []*jsonNode{}
	for i := range numFiles {
		if err := createFileLinks(i, workDir, graph); err != nil {
			return fmt.Errorf("unable to create graph: %s", err.Error())
		}

		node := &jsonNode{
			Id:        fmt.Sprintf("n%d", i),
			Neighbors: []string{},
		}

		for j := range numFiles {
			if graph.Edges[i][j] {
				node.Neighbors = append(node.Neighbors, fmt.Sprintf("n%d", j))
			}
		}
		nodes = append(nodes, node)
	}
	jsonStruct := &jsonGraph{
		Graph: nodes,
	}

	jsonText, err := json.Marshal(jsonStruct)
	if err != nil {
		return fmt.Errorf("unable to create graph: %s", err.Error())
	}

	filename := sourceDir + "_graph.json"
	err = os.WriteFile(filename, jsonText, 0777)
	if err != nil {
		return fmt.Errorf("unable to create graph: %s", err.Error())
	}

	if err := cleanupWorkingEnvironment(workDir); err != nil {
		return fmt.Errorf("unable to create graph: %s", err.Error())
	}

	return nil
}

func prepareWorkingEnvironment(sourceDir, workDir string) error {
	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf("Unable to prepare clean working environment: %s", err.Error())
	}

	sourceDirFs := os.DirFS(sourceDir)

	if err := os.CopyFS(workDir, sourceDirFs); err != nil {
		return fmt.Errorf("Unable to create working copy of vault: %s", err.Error())
	}

	log.Printf("created working copy of vault at: %s", workDir)

	return nil
}

func flattenWorkingEnvironment(numFiles int, workDir string, graph *graph) error {
	subdirectories := []string{}

	i := 0
	titles := make(NodeProperty, numFiles)
	paths := make(NodeProperty, numFiles)

	flatten := func(path string, location fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		originalPath := path
		originalFullPath := strings.Join([]string{workDir, originalPath}, "/")

		if !location.IsDir() {
			newPath := cleanStr(path)
			newFullPath := strings.Join([]string{workDir, newPath}, "/")

			if err := os.Rename(originalFullPath, newFullPath); err != nil {
				return err
			}

			titles[i] = strings.ReplaceAll(filepath.Base(originalPath), ".md", "")
			paths[i] = newPath
			i = i + 1
		} else {
			subdirectories = append(subdirectories, originalFullPath)
		}

		return nil
	}

	workingFs := os.DirFS(workDir)
	if err := fs.WalkDir(workingFs, ".", flatten); err != nil {
		return fmt.Errorf("Unable to flatten vault: %s", err.Error())
	}

	graph.NodeProperties[titleKey] = titles
	graph.NodeProperties[pathKey] = paths

	for _, d := range subdirectories {
		if err := os.RemoveAll(d); err != nil {
			return fmt.Errorf("Unable to remove subdirectory %s: %s", d, err.Error())
		}
	}

	return nil
}

func cleanupWorkingEnvironment(workDir string) error {
	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf("Unable to clean up working copy of vault: %s", err.Error())
	}

	log.Printf("cleanup complete")

	return nil
}

func createFileLinks(fileIndex int, workDir string, graph *graph) error {
	filePath := filepath.Join(workDir, graph.NodeProperties[pathKey][fileIndex])

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file: %s", err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineContainsLink := strings.Contains(line, "[[") && strings.Contains(line, "]]")

		if lineContainsLink {
			linkTitle := strings.Split(strings.Split(line, "[[")[1], "]]")[0]
			linkDestination := graph.firstIndexOf(titleKey, linkTitle)

			graph.Edges[fileIndex][linkDestination] = true
		}
	}

	return nil
}

func (graph *graph) firstIndexOf(key, value string) int {
	for i := range graph.NodeProperties[key] {
		if graph.NodeProperties[key][i] == value {
			return i
		}
	}

	return -1
}

func countFiles(path string) (int, error) {
	count := 0
	dir := os.DirFS(path)

	inc := func(path string, location fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		if !location.IsDir() {
			count = count + 1
		}

		return nil
	}

	if err := fs.WalkDir(dir, ".", inc); err != nil {
		return 0, err
	}

	return count, nil
}

func cleanStr(str string) string {
	str = strings.ReplaceAll(str, ",", "")
	str = strings.ReplaceAll(str, "'", "")
	str = strings.ReplaceAll(str, "-", "")
	str = strings.ReplaceAll(str, "  ", " ")
	str = strings.ReplaceAll(str, " ", "_")
	str = strings.ReplaceAll(str, "/", "_")
	str = strings.ToLower(str)

	return str
}
