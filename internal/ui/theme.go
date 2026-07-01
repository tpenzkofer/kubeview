package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme is a named colour palette.
type Theme struct {
	Name                                                            string
	Frame, Title, Dim, Text, Green, Yellow, Red, Cyan, Magenta, Sel string
	GradLo, GradMid, GradHi                                         [3]int
}

var themes = map[string]Theme{
	"tokyonight": {
		Name: "tokyonight", Frame: "#2a3550", Title: "#7aa2f7", Dim: "#5c6773", Text: "#c0caf5",
		Green: "#9ece6a", Yellow: "#e0af68", Red: "#f7768e", Cyan: "#2ac3de", Magenta: "#bb9af7", Sel: "#283457",
		GradLo: [3]int{0x6a, 0xd1, 0x4f}, GradMid: [3]int{0xe6, 0xc8, 0x4a}, GradHi: [3]int{0xf4, 0x4a, 0x5e},
	},
	"gruvbox": {
		Name: "gruvbox", Frame: "#504945", Title: "#83a598", Dim: "#928374", Text: "#ebdbb2",
		Green: "#b8bb26", Yellow: "#fabd2f", Red: "#fb4934", Cyan: "#8ec07c", Magenta: "#d3869b", Sel: "#3c3836",
		GradLo: [3]int{0xb8, 0xbb, 0x26}, GradMid: [3]int{0xfa, 0xbd, 0x2f}, GradHi: [3]int{0xfb, 0x49, 0x34},
	},
	"nord": {
		Name: "nord", Frame: "#434c5e", Title: "#88c0d0", Dim: "#616e88", Text: "#d8dee9",
		Green: "#a3be8c", Yellow: "#ebcb8b", Red: "#bf616a", Cyan: "#8fbcbb", Magenta: "#b48ead", Sel: "#3b4252",
		GradLo: [3]int{0xa3, 0xbe, 0x8c}, GradMid: [3]int{0xeb, 0xcb, 0x8b}, GradHi: [3]int{0xbf, 0x61, 0x6a},
	},
	"dracula": {
		Name: "dracula", Frame: "#44475a", Title: "#bd93f9", Dim: "#6272a4", Text: "#f8f8f2",
		Green: "#50fa7b", Yellow: "#f1fa8c", Red: "#ff5555", Cyan: "#8be9fd", Magenta: "#ff79c6", Sel: "#44475a",
		GradLo: [3]int{0x50, 0xfa, 0x7b}, GradMid: [3]int{0xf1, 0xfa, 0x8c}, GradHi: [3]int{0xff, 0x55, 0x55},
	},
	"mono": {
		Name: "mono", Frame: "#444444", Title: "#ffffff", Dim: "#6a6a6a", Text: "#d0d0d0",
		Green: "#9a9a9a", Yellow: "#c4c4c4", Red: "#f4f4f4", Cyan: "#bcbcbc", Magenta: "#c8c8c8", Sel: "#3a3a3a",
		GradLo: [3]int{0x60, 0x60, 0x60}, GradMid: [3]int{0xa6, 0xa6, 0xa6}, GradHi: [3]int{0xf2, 0xf2, 0xf2},
	},
}

// ThemeNames lists the available themes (for help / flag docs).
func ThemeNames() []string { return []string{"tokyonight", "gruvbox", "nord", "dracula", "mono"} }

// Palette, set by ApplyTheme.
var (
	colFrame, colTitle, colDim, colText            lipgloss.Color
	colGreen, colYellow, colRed, colCyan, colMagenta lipgloss.Color
	colSelBG                                       lipgloss.Color

	styTitle, styDim, styText, styHeader, stySelect lipgloss.Style
	frame                                           lipgloss.Style

	gradLo, gradMid, gradHi [3]int
	currentTheme            = "tokyonight"
)

// ApplyTheme switches the active palette and rebuilds derived styles. Unknown
// names fall back to tokyonight. Returns the resolved theme name.
func ApplyTheme(name string) string {
	t, ok := themes[name]
	if !ok {
		t = themes["tokyonight"]
	}
	currentTheme = t.Name
	colFrame = lipgloss.Color(t.Frame)
	colTitle = lipgloss.Color(t.Title)
	colDim = lipgloss.Color(t.Dim)
	colText = lipgloss.Color(t.Text)
	colGreen = lipgloss.Color(t.Green)
	colYellow = lipgloss.Color(t.Yellow)
	colRed = lipgloss.Color(t.Red)
	colCyan = lipgloss.Color(t.Cyan)
	colMagenta = lipgloss.Color(t.Magenta)
	colSelBG = lipgloss.Color(t.Sel)
	gradLo, gradMid, gradHi = t.GradLo, t.GradMid, t.GradHi

	styTitle = lipgloss.NewStyle().Foreground(colTitle).Bold(true)
	styDim = lipgloss.NewStyle().Foreground(colDim)
	styText = lipgloss.NewStyle().Foreground(colText)
	styHeader = lipgloss.NewStyle().Foreground(colCyan).Bold(true)
	stySelect = lipgloss.NewStyle().Background(colSelBG).Foreground(colText).Bold(true)
	frame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colFrame)
	return t.Name
}

func init() { ApplyTheme("tokyonight") }

// gradRGB interpolates the theme's green→yellow→red ramp across 0..1.
func gradRGB(frac float64) (int, int, int) {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	var a, b [3]int
	var t float64
	if frac < 0.5 {
		a, b, t = gradLo, gradMid, frac/0.5
	} else {
		a, b, t = gradMid, gradHi, (frac-0.5)/0.5
	}
	return a[0] + int(float64(b[0]-a[0])*t),
		a[1] + int(float64(b[1]-a[1])*t),
		a[2] + int(float64(b[2]-a[2])*t)
}

// gradColor smoothly interpolates green→yellow→red across 0..1 (true-colour).
func gradColor(frac float64) lipgloss.Color {
	r, g, b := gradRGB(frac)
	return rgb(r, g, b)
}

func rgb(r, g, b int) lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// dimRGB darkens an RGB triple by factor (0..1).
func dimRGB(r, g, b int, factor float64) lipgloss.Color {
	return rgb(int(float64(r)*factor), int(float64(g)*factor), int(float64(b)*factor))
}

func hexToRGB(c lipgloss.Color) [3]int {
	s := strings.TrimPrefix(string(c), "#")
	if len(s) != 6 {
		return [3]int{0xc0, 0xc0, 0xc0}
	}
	var out [3]int
	for i := 0; i < 3; i++ {
		v, _ := strconv.ParseInt(s[i*2:i*2+2], 16, 0)
		out[i] = int(v)
	}
	return out
}

// fadeColor returns a row text colour that fades from the theme's text colour at
// the top of a list toward its dim colour at the bottom (btop's proc-list look).
func fadeColor(pos, n int) lipgloss.Color {
	if n <= 1 {
		return colText
	}
	t := float64(pos) / float64(n-1)
	if t > 1 {
		t = 1
	}
	t *= 0.72
	hi, lo := hexToRGB(colText), hexToRGB(colDim)
	return rgb(
		hi[0]+int(float64(lo[0]-hi[0])*t),
		hi[1]+int(float64(lo[1]-hi[1])*t),
		hi[2]+int(float64(lo[2]-hi[2])*t),
	)
}

// loadColor keeps the discrete green/yellow/red for non-gradient uses.
func loadColor(frac float64) lipgloss.Color {
	switch {
	case frac >= 0.85:
		return colRed
	case frac >= 0.60:
		return colYellow
	default:
		return colGreen
	}
}

// statusColor colours a pod/container status string.
func statusColor(status string) lipgloss.Color {
	switch status {
	case "Running", "Completed", "Succeeded":
		return colGreen
	case "Pending", "ContainerCreating", "PodInitializing", "Terminating":
		return colYellow
	case "":
		return colDim
	default:
		// CrashLoopBackOff, Error, ImagePullBackOff, OOMKilled, Evicted, ...
		return colRed
	}
}
