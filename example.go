package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Metadata struct {
	Files []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Chunks []int  `json:"chunks"`
	} `json:"files"`
}

func readMetadata(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func readIndex(path string) (map[string][]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	index := make(map[string][]int)
	for {
		var idLen uint8
		if err := binary.Read(f, binary.LittleEndian, &idLen); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		idBytes := make([]byte, idLen)
		if _, err := f.Read(idBytes); err != nil {
			return nil, err
		}
		fileID := string(idBytes)

		var chunkCount uint8
		if err := binary.Read(f, binary.LittleEndian, &chunkCount); err != nil {
			return nil, err
		}

		chunks := make([]int, chunkCount)
		for i := 0; i < int(chunkCount); i++ {
			var folder uint8
			if err := binary.Read(f, binary.LittleEndian, &folder); err != nil {
				return nil, err
			}
			chunks[i] = int(folder)
		}

		index[fileID] = chunks
	}
	return index, nil
}

func reassembleFile(meta *Metadata, index map[string][]int, outputPath string) error {
	file := meta.Files[0] // For demo, use first file
	chunks := index[file.ID]

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for i, folder := range chunks {
		chunkPath := filepath.Join(fmt.Sprintf("%d", folder), fmt.Sprintf("chunk%d.bin", i+1))
		data, err := os.ReadFile(chunkPath)
		if err != nil {
			return err
		}
		if _, err := out.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func sample() {
	meta, err := readMetadata("metadata.json")
	if err != nil {
		panic(err)
	}

	index, err := readIndex("filedata.id")
	if err != nil {
		panic(err)
	}

	if err := reassembleFile(meta, index, "output.txt"); err != nil {
		panic(err)
	}

	fmt.Println("File reconstructed successfully!")
}
