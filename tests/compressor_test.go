package tests

import (
	"ImageProcessor/pipeline"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestStageCompressBytes(t *testing.T) {
	// read from file ss/19495431737_1156_2023_7_26_16_18_39_1690402719421980.png
	file, err := os.Open("ss/19495431737_1156_2023_7_26_16_18_39_1690402719421980.png")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	compressedImage, err := pipeline.StageCompressBytes(
		content,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(compressedImage) == 0 {
		t.Fatal("compressed image is empty")
	}
}
