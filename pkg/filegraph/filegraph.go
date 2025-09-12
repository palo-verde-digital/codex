package filegraph

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
)

const (
	inKey  = "in"
	outKey = "out"

	originalPathKey = "originalPaths"
	newPathKey      = "newPaths"
)

type NodeValue []int
type EdgeValue [][]int

type NodeProperty []string
type EdgeProperty [][]string

type G struct {
	Size  int      `json:"size"`
	Edges [][]bool `json:"edges"`

	NodeValues map[string]NodeValue `json:"nodeValues"`
	EdgeValues map[string]EdgeValue `json:"edgeValues"`

	NodeProperties map[string]NodeProperty `json:"nodeProperties"`
	EdgeProperties map[string]EdgeProperty `json:"edgeProperties"`
}

func newGraph(size int) *G {
	edges := make([][]bool, size)
	for i := range edges {
		edges[i] = make([]bool, size)
	}

	return &G{
		Size:  size,
		Edges: edges,

		NodeValues: make(map[string]NodeValue),
		EdgeValues: make(map[string]EdgeValue),

		NodeProperties: make(map[string]NodeProperty),
		EdgeProperties: make(map[string]EdgeProperty),
	}
}

func (g *G) String() string {
	jsonBytes, err := json.MarshalIndent(g, "", "	")
	if err != nil {
		return "{}"
	}

	return string(jsonBytes)
}

func Create(source string) (*G, error) {
	destination := source + "_tmp"

	if err := os.RemoveAll(destination); err != nil {
		return nil, fmt.Errorf("Unable to prepare clean working environment: %s", err.Error())
	}

	sourceFs := os.DirFS(source)

	if err := os.CopyFS(destination, sourceFs); err != nil {
		return nil, fmt.Errorf("Unable to create working copy of vault: %s", err.Error())
	}

	log.Printf("created working copy of vault at: %s", destination)

	numFiles, err := countFiles(source)
	if err != nil {
		return nil, fmt.Errorf("unable to count files: %s", err.Error())
	}

	g := newGraph(numFiles)

	subdirectories := []string{}

	i := 0
	originalPaths := make(NodeProperty, numFiles)
	newPaths := make(NodeProperty, numFiles)

	flatten := func(path string, location fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		originalPath := path
		originalFullPath := strings.Join([]string{destination, originalPath}, "/")

		if !location.IsDir() {
			newPath := cleanStr(path)
			newFullPath := strings.Join([]string{destination, newPath}, "/")

			if err := os.Rename(originalFullPath, newFullPath); err != nil {
				return err
			}

			g.Edges[i][i] = true
			originalPaths[i] = originalPath
			newPaths[i] = newPath
			i = i + 1
		} else {
			subdirectories = append(subdirectories, originalFullPath)
		}

		return nil
	}

	workingFs := os.DirFS(destination)
	if err := fs.WalkDir(workingFs, ".", flatten); err != nil {
		return nil, fmt.Errorf("Unable to flatten vault: %s", err.Error())
	}

	g.NodeProperties[originalPathKey] = originalPaths
	g.NodeProperties[newPathKey] = newPaths

	for _, d := range subdirectories {
		if err := os.RemoveAll(d); err != nil {
			return nil, fmt.Errorf("Unable to remove subdirectory %s: %s", d, err.Error())
		}
	}

	log.Printf("flattened %d subdirectories. working vault created successfully", len(subdirectories))

	if err := os.RemoveAll(destination); err != nil {
		return nil, fmt.Errorf("Unable to clean up working copy of vault: %s", err.Error())
	}

	log.Printf("cleanup complete")

	return g, nil
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
