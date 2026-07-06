package dsp

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const ExpectedSampleRate = 44100

func ReadMonoWAV(filepath string) ([]float64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	header := make([]byte, 44)
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, fmt.Errorf("failed to read WAV header: %w", err)
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a valid WAVE file")
	}

	numChannels := binary.LittleEndian.Uint16(header[22:24])
	if numChannels != 1 {
		return nil, fmt.Errorf("audio must be Mono")
	}

	dataSize := binary.LittleEndian.Uint32(header[40:44])
	numSamples := dataSize / 2
	rawData := make([]byte, dataSize)
	if _, err := io.ReadFull(file, rawData); err != nil && err != io.EOF {
		return nil, err
	}

	samples := make([]float64, numSamples)
	for i := 0; i < int(numSamples); i++ {
		offset := i * 2
		sampleInt16 := int16(binary.LittleEndian.Uint16(rawData[offset : offset+2]))
		samples[i] = float64(sampleInt16) / 32768.0
	}
	return samples, nil
}