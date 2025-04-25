package main

import (
	"flag"
	"line2tg/apng2gif"
	"log"
	"os"
)

func ConvertFile(inPath, outPath string) error {
	in, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	return apng2gif.Convert(in, out)
}

func main() {
	// Define command-line flags
	inputPath := flag.String("input", "", "Path to input APNG file (required)")
	outputPath := flag.String("output", "output.gif", "Path to output GIF file")

	// Parse flags
	flag.Parse()

	// Check if input path is provided
	if *inputPath == "" {
		flag.Usage()
		log.Fatal("Error: input file path is required")
	}

	// Perform conversion
	err := ConvertFile(*inputPath, *outputPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Conversion completed successfully!")
}
