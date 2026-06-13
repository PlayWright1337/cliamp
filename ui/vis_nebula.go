package ui

import (
	"math"
	"strings"
)

func (v *Visualizer) renderNebula(bands []float64) string {
	rows := v.Rows
	dotRows := rows * 4
	dotCols := PanelWidth * 2
	if dotRows < 4 || dotCols < 8 {
		return strings.Repeat("\n", max(0, rows-1))
	}

	grid := brailleGrid{}
	grid.ensure(dotRows, dotCols)
	bass := bandAvg(bands, 0, len(bands)/3)
	mid := bandAvg(bands, len(bands)/3, 2*len(bands)/3)
	high := bandAvg(bands, 2*len(bands)/3, len(bands))
	cx := float64(dotCols-1) / 2
	cy := float64(dotRows-1) / 2
	t := float64(v.frame)
	count := 42 + int((bass+mid+high)*22)
	for i := 0; i < count; i++ {
		seed := float64(i + 1)
		radius := math.Mod(seed*7.37+t*(0.035+high*0.05), 1) * float64(min(dotRows, dotCols)) * 0.58
		angle := seed*2.399963 + t*(0.006+bass*0.020) + math.Sin(t*0.012+seed)*mid*0.9
		x := int(cx + math.Cos(angle)*radius*1.55)
		y := int(cy + math.Sin(angle)*radius*0.72)
		tier := int8(1)
		switch {
		case radius < float64(min(dotRows, dotCols))*0.16 || bass > 0.72:
			tier = 3
		case mid > 0.35 || radius < float64(min(dotRows, dotCols))*0.32:
			tier = 2
		}
		grid.set(x, y, tier)
		if high > 0.42 && i%3 == 0 {
			grid.set(x+1, y, tier)
		}
		if bass > 0.55 && i%5 == 0 {
			grid.drawLine(int(cx), int(cy), x, y, 1)
		}
	}
	return grid.render(rows)
}
