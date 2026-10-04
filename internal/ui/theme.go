package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

var (
	colorOcean        = lipgloss.Color("#071E26")
	colorPanel        = lipgloss.Color("#0C2832")
	colorStatus       = lipgloss.Color("#04161C")
	colorSand         = lipgloss.Color("#F4E4C4")
	colorSandDim      = lipgloss.Color("#B7A88A")
	colorLagoon       = lipgloss.Color("#1F8F86")
	colorLagoonBright = lipgloss.Color("#3DDECE")
	colorCoral        = lipgloss.Color("#FF6F61")
	colorPalm         = lipgloss.Color("#7DCE82")
	colorSelect       = lipgloss.Color("#1B5E5A")
	colorSelectDim    = lipgloss.Color("#14343C")
	colorBorderDim    = lipgloss.Color("#24555C")
)
