# Welcome to asefile

Asefile is a library for loading the aseprite file format. Included is an example of how to use the library, and then render a frame from aseprite using ebiten.

- Not all features have been tested yet - eg encoding back into a file
- Files where the image is encoded in different ways eg, raw pixel data as opposed to zlib compressed
- Etc

# Loading and rendering a frame
The below code is a brief example of how you can load and render a frame to an ebiten image
```go
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
```

# Run the example
If you clone the repository then
`go run example/main.go`

![result](chica.png)