package main

import (
	"fmt"
	"os"

	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

// LoadTTF loads a TTF font from a file and returns a text.Atlas
func LoadTTF(path string, size float64) (*text.Atlas, error) {
	ttfData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read TTF file: %w", err)
	}

	ttfFont, err := opentype.Parse(ttfData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TTF file: %w", err)
	}

	face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}

	return text.NewAtlas(face, text.ASCII), nil
}

// LoadBasicFont loads the basic font and returns a text.Atlas
func LoadBasicFont() *text.Atlas {
	return text.NewAtlas(basicfont.Face7x13, text.ASCII)
}

// DrawText draws text on the window at the specified position
func DrawText(win *pixelgl.Window, atlas *text.Atlas, content string, pos pixel.Vec, scale float64, color color.RGBA) {
	txt := text.New(pos, atlas)
	txt.Color = color
	fmt.Fprintln(txt, content)
	txt.Draw(win, pixel.IM.Scaled(txt.Orig, scale))
}
