package x

import (
	"fmt"
	"strings"
)

type progressLine struct {
	maxWidth int
}

func (p *progressLine) Update(format string, args ...any) {
	p.print(false, format, args...)
}

func (p *progressLine) Finish(format string, args ...any) {
	p.print(true, format, args...)
	p.maxWidth = 0
}

func (p *progressLine) print(done bool, format string, args ...any) {
	line := fmt.Sprintf("[xdl] "+format, args...)
	if len(line) < p.maxWidth {
		line += strings.Repeat(" ", p.maxWidth-len(line))
	} else {
		p.maxWidth = len(line)
	}

	if done {
		fmt.Printf("\r%s\n", line)
		return
	}

	fmt.Printf("\r%s", line)
}

func renderProgressBar(current, total, width int) string {
	if width <= 0 {
		width = 20
	}
	if total <= 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}

	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if current > 0 && filled == 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}

	return "[" + strings.Repeat("=", filled) + strings.Repeat("-", width-filled) + "]"
}
