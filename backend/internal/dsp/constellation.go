package dsp

const (
	FreqNeighborhood = 10
	TimeNeighborhood = 10
	MinAmplitude     = 2.0
)

type Peak struct {
	TimeFrame int
	FreqBin   int
	Amplitude float64
}

func ExtractPeaks(spectrogram [][]float64) []Peak {
	var peaks []Peak
	timeFrames := len(spectrogram)
	if timeFrames == 0 {
		return peaks
	}
	freqBins := len(spectrogram[0])

	for t := 0; t < timeFrames; t++ {
		for f := 0; f < freqBins; f++ {
			amplitude := spectrogram[t][f]
			if amplitude < MinAmplitude {
				continue
			}
			isMax := true
			tMin := max(0, t-TimeNeighborhood)
			tMax := min(timeFrames-1, t+TimeNeighborhood)
			fMin := max(0, f-FreqNeighborhood)
			fMax := min(freqBins-1, f+FreqNeighborhood)

			for nt := tMin; nt <= tMax && isMax; nt++ {
				for nf := fMin; nf <= fMax; nf++ {
					if nt == t && nf == f {
						continue
					}
					if spectrogram[nt][nf] >= amplitude {
						isMax = false
						break
					}
				}
			}
			if isMax {
				peaks = append(peaks, Peak{TimeFrame: t, FreqBin: f, Amplitude: amplitude})
			}
		}
	}
	return peaks
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}