package dsp

import (
	"math/cmplx"
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

const (
	ChunkSize = 4096
	Overlap   = 2048
	StepSize  = ChunkSize - Overlap
)

func GenerateSpectrogram(audio []float64) [][]float64 {
	var spectrogram [][]float64
	hammingWin := window.Hamming(ChunkSize)

	for start := 0; start+ChunkSize <= len(audio); start += StepSize {
		chunk := audio[start : start+ChunkSize]
		windowedChunk := make([]float64, ChunkSize)
		for i := 0; i < ChunkSize; i++ {
			windowedChunk[i] = chunk[i] * hammingWin[i]
		}
		fftResult := fft.FFTReal(windowedChunk)
		halfSize := ChunkSize / 2
		magnitudes := make([]float64, halfSize)
		for i := 0; i < halfSize; i++ {
			magnitudes[i] = cmplx.Abs(fftResult[i])
		}
		spectrogram = append(spectrogram, magnitudes)
	}
	return spectrogram
}