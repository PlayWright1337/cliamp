package ui

import (
	"strings"
)

func (v *Visualizer) renderVU(bands []float64) string {
	rows := v.Rows
	if rows <= 0 {
		return ""
	}
	bass := bandAvg(bands, 0, len(bands)/3)
	mid := bandAvg(bands, len(bands)/3, 2*len(bands)/3)
	high := bandAvg(bands, 2*len(bands)/3, len(bands))
	left := bass*0.55 + mid*0.30 + high*0.15
	right := bass*0.25 + mid*0.35 + high*0.40
	if left > 1 {
		left = 1
	}
	if right > 1 {
		right = 1
	}

	meterRows := rows
	if meterRows > 4 {
		meterRows = 4
	}
	topPad := (rows - meterRows) / 2
	lines := make([]string, rows)
	for i := range lines {
		lines[i] = strings.Repeat(" ", PanelWidth)
	}
	for r := 0; r < meterRows; r++ {
		level := left
		if r%2 == 1 {
			level = right
		}
		width := int(level * float64(PanelWidth))
		if width < 0 {
			width = 0
		}
		if width > PanelWidth {
			width = PanelWidth
		}
		var sb, run strings.Builder
		tag := -1
		for c := 0; c < PanelWidth; c++ {
			norm := float64(c) / float64(max(1, PanelWidth-1))
			nextTag := specTag(norm)
			if nextTag != tag {
				flushStyleRun(&sb, &run, tag)
				tag = nextTag
			}
			if c < width {
				if (c+int(v.frame/3))%8 == 7 {
					run.WriteRune('▌')
				} else {
					run.WriteRune('█')
				}
			} else if c%8 == 7 {
				run.WriteRune('·')
			} else {
				run.WriteRune(' ')
			}
		}
		flushStyleRun(&sb, &run, tag)
		lines[topPad+r] = sb.String()
	}
	return strings.Join(lines, "\n")
}
