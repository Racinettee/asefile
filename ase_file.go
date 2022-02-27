package asefile

import (
	"io"
	"os"
)

type AsepriteFile struct {
	Header AsepriteHeader
	Frames []AsepriteFrame
}

func (aseFile *AsepriteFile) Decode(r io.Reader) error {
	aseFile.Header.Decode(r)
	aseFile.Frames = make([]AsepriteFrame, aseFile.Header.Frames)
	for x := range aseFile.Frames {
		aseFile.Frames[x].parentHeader = &aseFile.Header
		err := aseFile.Frames[x].Decode(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (aseFile *AsepriteFile) Encode(w io.Writer) {
	aseFile.Header.Encode(w)

	aseFile.Frames = make([]AsepriteFrame, aseFile.Header.Frames)

	for _, frame := range aseFile.Frames {
		frame.Encode(w)
	}
}

func (aseFile *AsepriteFile) DecodeFile(fName string) error {
	spriteFile, err := os.Open(fName)
	if err != nil {
		return err
	}
	return aseFile.Decode(spriteFile)
}
