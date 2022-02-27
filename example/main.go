package main

import (
	"image/color"
	"log"

	"github.com/Racinettee/asefile"
	ebi "github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 240
	screenHeight = 240
)

type Game struct {
	aseImg *ebi.Image
}

func (g *Game) Init() {
	var aseFile asefile.AsepriteFile
	if err := aseFile.DecodeFile("example/Chica.aseprite"); err != nil {
		log.Fatal(err)
	}
	g.aseImg = ebi.NewImage(int(aseFile.Header.WidthInPixels), int(aseFile.Header.HeightInPixels))
	for _, cel := range aseFile.Frames[0].Cels {
		dat := cel.RawCelData
		w, h := cel.WidthInPix, cel.HeightInPix
		offset := 0
		for y := 0; y < int(h); y += 1 {
			for x := 0; x < int(w); x, offset = x+1, offset+4 {
				col := color.RGBA{dat[offset], dat[offset+1], dat[offset+2], dat[offset+3]}
				g.aseImg.Set(int(cel.X)+x, int(cel.Y)+y, col)
			}
		}
	}
}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebi.Image) {
	screen.Fill(color.RGBA{100, 237, 149, 255})
	op := ebi.DrawImageOptions{}
	op.GeoM.Translate(screenWidth/2, screenHeight/2)
	screen.DrawImage(g.aseImg, &op)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebi.SetWindowSize(screenWidth*2, screenHeight*2)
	ebi.SetWindowTitle("Aseprite File Example")
	g := &Game{}
	g.Init()
	if err := ebi.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
