package asefile

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type AsepriteCodec interface {
	Decode(r io.Reader) error
	Encode(w io.Writer)
}

type AsepriteUserDatHolder interface {
	AddUserData(AsepriteUserDataChunk2020)
}

var ble = binary.LittleEndian

func DecodeAseString(r io.Reader) string {
	var len uint16
	binary.Read(r, ble, &len)
	buff := make([]byte, len)
	binary.Read(r, ble, &buff)
	return string(buff)
}

func EncodeAseString(w io.Writer, str string) {
	len := uint16(len(str))
	binary.Write(w, ble, &len)
	buff := []byte(str)
	binary.Write(w, ble, &buff)
}

func (aseHeader *AsepriteHeader) Decode(r io.Reader) error {
	binary.Read(r, ble, &aseHeader.FileSize)
	binary.Read(r, ble, &aseHeader.MagicNumber)

	if aseHeader.MagicNumber != 0xA5E0 {
		return fmt.Errorf("header magic number incorrect")
	}

	binary.Read(r, ble, &aseHeader.Frames)
	binary.Read(r, ble, &aseHeader.WidthInPixels)
	binary.Read(r, ble, &aseHeader.HeightInPixels)
	binary.Read(r, ble, &aseHeader.ColorDepth)
	binary.Read(r, ble, &aseHeader.Flags)
	binary.Read(r, ble, &aseHeader.Speed)
	binary.Read(r, ble, &aseHeader.ignore1)
	binary.Read(r, ble, &aseHeader.ignore2)
	binary.Read(r, ble, &aseHeader.PaletteEntry)
	binary.Read(r, ble, &aseHeader.ignore3)
	binary.Read(r, ble, &aseHeader.NumberOfColors)
	binary.Read(r, ble, &aseHeader.PixelWidth)
	binary.Read(r, ble, &aseHeader.PixelHeight)
	binary.Read(r, ble, &aseHeader.XPositionOfGrid)
	binary.Read(r, ble, &aseHeader.YPositionOfGrid)
	binary.Read(r, ble, &aseHeader.GridWidth)
	binary.Read(r, ble, &aseHeader.GridHeight)
	binary.Read(r, ble, &aseHeader.reserved)
	return nil
}

func (aseHeader *AsepriteHeader) Encode(w io.Writer) {
	binary.Write(w, ble, &aseHeader.FileSize)
	binary.Write(w, ble, &aseHeader.MagicNumber)
	binary.Write(w, ble, &aseHeader.Frames)
	binary.Write(w, ble, &aseHeader.WidthInPixels)
	binary.Write(w, ble, &aseHeader.HeightInPixels)
	binary.Write(w, ble, &aseHeader.ColorDepth)
	binary.Write(w, ble, &aseHeader.Flags)
	binary.Write(w, ble, &aseHeader.Speed)
	binary.Write(w, ble, &aseHeader.ignore1)
	binary.Write(w, ble, &aseHeader.ignore2)
	binary.Write(w, ble, &aseHeader.PaletteEntry)
	binary.Write(w, ble, &aseHeader.ignore3)
	binary.Write(w, ble, &aseHeader.NumberOfColors)
	binary.Write(w, ble, &aseHeader.PixelWidth)
	binary.Write(w, ble, &aseHeader.PixelHeight)
	binary.Write(w, ble, &aseHeader.XPositionOfGrid)
	binary.Write(w, ble, &aseHeader.YPositionOfGrid)
	binary.Write(w, ble, &aseHeader.GridWidth)
	binary.Write(w, ble, &aseHeader.GridHeight)
	binary.Write(w, ble, &aseHeader.reserved)
}

func (aseFrame *AsepriteFrame) Decode(r io.Reader) error {
	binary.Read(r, ble, &aseFrame.BytesThisFrame)
	binary.Read(r, ble, &aseFrame.MagicNumber)

	if aseFrame.MagicNumber != 0xF1FA {
		return fmt.Errorf("frame magic number incorrect")
	}

	binary.Read(r, ble, &aseFrame.ChunksThisFrame)
	binary.Read(r, ble, &aseFrame.FrameDurationMilliseconds)
	binary.Read(r, ble, &aseFrame.reserved)
	binary.Read(r, ble, &aseFrame.ChunksThisFrameExt)
	//
	// Load n-amount of chunks
	aseFrame.OldPalettes0004 = make([]AsepriteOldPaletteChunk0004, 0)
	aseFrame.OldPalettes0011 = make([]AsepritePaletteChunk0011, 0)
	aseFrame.Layers = make([]AsepriteLayerChunk2004, 0)
	aseFrame.Cels = make([]AsepriteCelChunk2005, 0)
	aseFrame.ColorProfiles = make([]AsepriteColorProfileChunk2007, 0)
	aseFrame.Palettes = make([]AsepritePaletteChunk2019, 0)
	aseFrame.Slices = make([]AsepriteSliceChunk2022, 0)

	loadChunks := 0
	if aseFrame.ChunksThisFrameExt == 0 {
		loadChunks = int(aseFrame.ChunksThisFrame)
	} else {
		loadChunks = int(aseFrame.ChunksThisFrameExt)
	}
	var lastUserdatHolder AsepriteUserDatHolder
	read := 0
	for x := 0; x < loadChunks; x += 1 {
		var chunkSize uint32
		var chunkType uint16
		binary.Read(r, ble, &chunkSize)
		binary.Read(r, ble, &chunkType)

		switch chunkType {
		case 0x0004:
			var oldPalette0004 AsepriteOldPaletteChunk0004
			oldPalette0004.Decode(r)
			aseFrame.OldPalettes0004 = append(aseFrame.OldPalettes0004, oldPalette0004)
			read += 1
		case 0x0011:
			var oldPalette0011 AsepritePaletteChunk0011
			oldPalette0011.Decode(r)
			aseFrame.OldPalettes0011 = append(aseFrame.OldPalettes0011, oldPalette0011)
			read += 1
		case 0x2004:
			var layer AsepriteLayerChunk2004
			layer.Decode(r)
			aseFrame.Layers = append(aseFrame.Layers, layer)
			lastUserdatHolder = &aseFrame.Layers[len(aseFrame.Layers)-1]
			read += 1
		case 0x2005:
			var cel AsepriteCelChunk2005
			cel.parentHeader = aseFrame.parentHeader
			cel.chunkSize = chunkSize
			cel.Decode(r)
			aseFrame.Cels = append(aseFrame.Cels, cel)
			read += 1
		case 0x2007:
			var colProfile AsepriteColorProfileChunk2007
			colProfile.Decode(r)
			aseFrame.ColorProfiles = append(aseFrame.ColorProfiles, colProfile)
			read += 1
		case 0x2018:
			aseFrame.Tags.Decode(r)
			lastUserdatHolder = &aseFrame.Tags
			read += 1
		case 0x2019:
			var palette AsepritePaletteChunk2019
			palette.Decode(r)
			aseFrame.Palettes = append(aseFrame.Palettes, palette)
			read += 1
		case 0x2020:
			var userDat AsepriteUserDataChunk2020
			userDat.Decode(r)
			if lastUserdatHolder != nil {
				lastUserdatHolder.AddUserData(userDat)
			}
			read += 1
		case 0x2022:
			var sliceDat AsepriteSliceChunk2022
			sliceDat.Decode(r)
			aseFrame.Slices = append(aseFrame.Slices, sliceDat)
			read += 1
		default:
			log.Printf("Unused chunk type: %X\n", chunkType)
		}
	}
	if read != loadChunks {
		return fmt.Errorf("did not read expected amount of chunks")
	}
	return nil
}

func (aseFrame AsepriteFrame) Encode(w io.Writer) {
	binary.Write(w, ble, &aseFrame.BytesThisFrame)
	binary.Write(w, ble, &aseFrame.MagicNumber)
	binary.Write(w, ble, &aseFrame.ChunksThisFrame)
	binary.Write(w, ble, &aseFrame.FrameDurationMilliseconds)
	binary.Write(w, ble, &aseFrame.reserved)
	binary.Write(w, ble, &aseFrame.ChunksThisFrameExt)
	//
	// Write n-amount of chunks
}

func (asePaletteChunk *AsepriteOldPaletteChunk0004) Decode(r io.Reader) {
	binary.Read(r, ble, &asePaletteChunk.NumberOfPackets)
	asePaletteChunk.Packets = make([]AsepriteOldPaletteChunk0004Packet, asePaletteChunk.NumberOfPackets)
	for x := 0; x < int(asePaletteChunk.NumberOfPackets); x += 1 {
		binary.Read(r, ble, &asePaletteChunk.Packets[x].NumPalletteEntriesToSkip)
		binary.Read(r, ble, &asePaletteChunk.Packets[x].NumColorsInThePacket)
		asePaletteChunk.Packets[x].Colors = make([]AsepriteRGB24, asePaletteChunk.Packets[x].NumColorsInThePacket)
		for y := 0; y < int(asePaletteChunk.Packets[x].NumColorsInThePacket); y += 1 {
			binary.Read(r, ble, &asePaletteChunk.Packets[x].Colors[y].R)
			binary.Read(r, ble, &asePaletteChunk.Packets[x].Colors[y].G)
			binary.Read(r, ble, &asePaletteChunk.Packets[x].Colors[y].B)
		}
	}
}

func (asePaletteChunk AsepriteOldPaletteChunk0004) Encode(w io.Writer) {
	binary.Write(w, ble, &asePaletteChunk.NumberOfPackets)
	for x := 0; x < int(asePaletteChunk.NumberOfPackets); x += 1 {
		binary.Write(w, ble, &asePaletteChunk.Packets[x].NumPalletteEntriesToSkip)
		binary.Write(w, ble, &asePaletteChunk.Packets[x].NumColorsInThePacket)
		for y := 0; y < int(asePaletteChunk.Packets[x].NumColorsInThePacket); y += 1 {
			binary.Write(w, ble, &asePaletteChunk.Packets[x].Colors[y].R)
			binary.Write(w, ble, &asePaletteChunk.Packets[x].Colors[y].G)
			binary.Write(w, ble, &asePaletteChunk.Packets[x].Colors[y].B)
		}
	}
}

func (asePaletteChunk *AsepritePaletteChunk0011) Decode(r io.Reader) {
	binary.Read(r, ble, &asePaletteChunk.NumberOfPackets)
	for x := 0; x < int(asePaletteChunk.NumberOfPackets); x += 1 {
		binary.Read(r, ble, &asePaletteChunk.Packets[x].NumPalletteEntriesToSkip)
		binary.Read(r, ble, &asePaletteChunk.Packets[x].NumColorsInThePacket)
		asePaletteChunk.Packets[x].Colors = make([]AsepriteRGB24, asePaletteChunk.Packets[x].NumColorsInThePacket)
		for y := 0; y < int(asePaletteChunk.Packets[x].NumColorsInThePacket); y += 1 {
			binary.Read(r, ble, &asePaletteChunk.Packets[x].Colors[y].R)
			binary.Read(r, ble, &asePaletteChunk.Packets[x].Colors[y].G)
			binary.Read(r, ble, &asePaletteChunk.Packets[x].Colors[y].B)
		}
	}
}

func (asePaletteChunk AsepritePaletteChunk0011) Encode(w io.Writer) {
	binary.Write(w, ble, &asePaletteChunk.NumberOfPackets)
	for x := 0; x < int(asePaletteChunk.NumberOfPackets); x += 1 {
		binary.Write(w, ble, &asePaletteChunk.Packets[x].NumPalletteEntriesToSkip)
		binary.Write(w, ble, &asePaletteChunk.Packets[x].NumColorsInThePacket)
		for y := 0; y < int(asePaletteChunk.Packets[x].NumColorsInThePacket); y += 1 {
			binary.Write(w, ble, &asePaletteChunk.Packets[x].Colors[y].R)
			binary.Write(w, ble, &asePaletteChunk.Packets[x].Colors[y].G)
			binary.Write(w, ble, &asePaletteChunk.Packets[x].Colors[y].B)
		}
	}
}

func (aseLayerChunk *AsepriteLayerChunk2004) Decode(r io.Reader) error {
	binary.Read(r, ble, &aseLayerChunk.Flags)
	binary.Read(r, ble, &aseLayerChunk.LayerType)
	binary.Read(r, ble, &aseLayerChunk.LayerChildLevel)
	binary.Read(r, ble, &aseLayerChunk.DefLayerWidthPixels)
	binary.Read(r, ble, &aseLayerChunk.DefLayerHeightPixels)
	binary.Read(r, ble, &aseLayerChunk.BlendMode)
	binary.Read(r, ble, &aseLayerChunk.Opacity)
	binary.Read(r, ble, &aseLayerChunk.forFuture)
	aseLayerChunk.LayerName = DecodeAseString(r)
	if aseLayerChunk.LayerType == 2 {
		binary.Read(r, ble, &aseLayerChunk.TilesetIndex)
	}
	return nil
}

func (aseLayerChunk AsepriteLayerChunk2004) Encode(w io.Writer) {
	binary.Write(w, ble, &aseLayerChunk.Flags)
	binary.Write(w, ble, &aseLayerChunk.LayerType)
	binary.Write(w, ble, &aseLayerChunk.LayerChildLevel)
	binary.Write(w, ble, &aseLayerChunk.DefLayerWidthPixels)
	binary.Write(w, ble, &aseLayerChunk.DefLayerHeightPixels)
	binary.Write(w, ble, &aseLayerChunk.BlendMode)
	binary.Write(w, ble, &aseLayerChunk.Opacity)
	binary.Write(w, ble, &aseLayerChunk.forFuture)
	EncodeAseString(w, aseLayerChunk.LayerName)
	if aseLayerChunk.LayerType == 2 {
		binary.Write(w, ble, &aseLayerChunk.TilesetIndex)
	}
}

func (aseCelChunk *AsepriteCelChunk2005) Decode(r io.Reader) {
	binary.Read(r, ble, &aseCelChunk.LayerIndex)
	binary.Read(r, ble, &aseCelChunk.X)
	binary.Read(r, ble, &aseCelChunk.Y)
	binary.Read(r, ble, &aseCelChunk.OpacityLevel)
	binary.Read(r, ble, &aseCelChunk.CelType)
	binary.Read(r, ble, &aseCelChunk.future)
	switch aseCelChunk.CelType {
	case 0:
		binary.Read(r, ble, &aseCelChunk.WidthInPix)
		binary.Read(r, ble, &aseCelChunk.HeightInPix)
		bytesToAlloc := int(aseCelChunk.WidthInPix) * int(aseCelChunk.HeightInPix)
		switch aseCelChunk.parentHeader.ColorDepth {
		case 32:
			bytesToAlloc *= 4
		case 16:
			bytesToAlloc *= 2
		case 8:
			break
		}
		aseCelChunk.RawPixData = make([]byte, bytesToAlloc)
		binary.Read(r, ble, &aseCelChunk.RawPixData)
	case 1:
		binary.Read(r, ble, &aseCelChunk.FramePosToLinkWith)
	case 2:
		binary.Read(r, ble, &aseCelChunk.WidthInPix)
		binary.Read(r, ble, &aseCelChunk.HeightInPix)
		bytesToRead := aseCelChunk.chunkSize - 26
		bytesBuff := make([]byte, bytesToRead)
		binary.Read(r, ble, bytesBuff)
		zreader, err := zlib.NewReader(bytes.NewReader(bytesBuff))
		if err != nil {
			log.Println(err)
		}
		byteBuff := bytes.NewBuffer([]byte{})
		io.Copy(byteBuff, zreader)
		aseCelChunk.RawCelData = byteBuff.Bytes()
	case 3:
		binary.Read(r, ble, &aseCelChunk.WidthInTiles)
		binary.Read(r, ble, &aseCelChunk.HeightInTiles)
		binary.Read(r, ble, &aseCelChunk.BitsPerTile)
		binary.Read(r, ble, &aseCelChunk.BitMaskForTileID)
		binary.Read(r, ble, &aseCelChunk.BitMaskForXFlip)
		binary.Read(r, ble, &aseCelChunk.BitMaskForYFlip)
		binary.Read(r, ble, &aseCelChunk.BitMaskFor90CWRot)
		binary.Read(r, ble, &aseCelChunk.reserved)
		zreader, err := zlib.NewReader(r)
		if err != nil {
			log.Println(err)
		}
		byteBuff := bytes.NewBuffer([]byte{})
		io.Copy(byteBuff, zreader)
		aseCelChunk.Tiles = byteBuff.Bytes()
	}
}

func (aseCelChunk *AsepriteCelChunk2005) Encode(w io.Writer) {
	binary.Write(w, ble, &aseCelChunk.LayerIndex)
	binary.Write(w, ble, &aseCelChunk.X)
	binary.Write(w, ble, &aseCelChunk.Y)
	binary.Write(w, ble, &aseCelChunk.OpacityLevel)
	binary.Write(w, ble, &aseCelChunk.CelType)
	binary.Write(w, ble, &aseCelChunk.future)
	switch aseCelChunk.CelType {
	case 0:
		binary.Write(w, ble, &aseCelChunk.WidthInPix)
		binary.Write(w, ble, &aseCelChunk.HeightInPix)
		bytesToAlloc := int(aseCelChunk.WidthInPix) * int(aseCelChunk.HeightInPix)
		switch aseCelChunk.parentHeader.ColorDepth {
		case 32:
			bytesToAlloc *= 4
		case 16:
			bytesToAlloc *= 2
		case 8:
			break
		}
		binary.Write(w, ble, &aseCelChunk.RawPixData)
	case 1:
		binary.Write(w, ble, &aseCelChunk.FramePosToLinkWith)
	case 2:
		binary.Write(w, ble, &aseCelChunk.WidthInPix)
		binary.Write(w, ble, &aseCelChunk.HeightInPix)
		var byteBuff bytes.Buffer
		zwriter := zlib.NewWriter(&byteBuff)
		zwriter.Write(aseCelChunk.RawCelData)
		w.Write(byteBuff.Bytes())
	case 3:
		binary.Write(w, ble, &aseCelChunk.WidthInTiles)
		binary.Write(w, ble, &aseCelChunk.HeightInTiles)
		binary.Write(w, ble, &aseCelChunk.BitsPerTile)
		binary.Write(w, ble, &aseCelChunk.BitMaskForTileID)
		binary.Write(w, ble, &aseCelChunk.BitMaskForXFlip)
		binary.Write(w, ble, &aseCelChunk.BitMaskForYFlip)
		binary.Write(w, ble, &aseCelChunk.BitMaskFor90CWRot)
		binary.Write(w, ble, &aseCelChunk.reserved)
		var byteBuff bytes.Buffer
		zwriter := zlib.NewWriter(&byteBuff)
		zwriter.Write(aseCelChunk.Tiles)
		w.Write(byteBuff.Bytes())
	}
}

func (aseCelExtra *AsepriteCelExtraChunk2006) Decode(r io.Reader) {
	binary.Read(r, ble, &aseCelExtra.Flags)
	binary.Read(r, ble, &aseCelExtra.PreciseX)
	binary.Read(r, ble, &aseCelExtra.PreciseY)
	binary.Read(r, ble, &aseCelExtra.WidthCelInSprite)
	binary.Read(r, ble, &aseCelExtra.HeightCelInSprite)
	binary.Read(r, ble, &aseCelExtra.futureUse)
}

func (aseCelExtra *AsepriteCelExtraChunk2006) Encode(w io.Writer) {
	binary.Write(w, ble, &aseCelExtra.Flags)
	binary.Write(w, ble, &aseCelExtra.PreciseX)
	binary.Write(w, ble, &aseCelExtra.PreciseY)
	binary.Write(w, ble, &aseCelExtra.WidthCelInSprite)
	binary.Write(w, ble, &aseCelExtra.HeightCelInSprite)
	binary.Write(w, ble, &aseCelExtra.futureUse)
}

func (aseColProfile *AsepriteColorProfileChunk2007) Decode(r io.Reader) {
	binary.Read(r, ble, &aseColProfile.Type)
	binary.Read(r, ble, &aseColProfile.Flags)
	binary.Read(r, ble, &aseColProfile.FixedGamma)
	binary.Read(r, ble, &aseColProfile.reserved)
	if aseColProfile.Type == 2 {
		binary.Read(r, ble, &aseColProfile.ICCProfileDatLen)
		aseColProfile.ICCProfileDat = make([]byte, aseColProfile.ICCProfileDatLen)
		binary.Read(r, ble, &aseColProfile.ICCProfileDat)
	}
}

func (aseColProfile AsepriteColorProfileChunk2007) Encode(w io.Writer) {
	binary.Write(w, ble, &aseColProfile.Type)
	binary.Write(w, ble, &aseColProfile.Flags)
	binary.Write(w, ble, &aseColProfile.FixedGamma)
	binary.Write(w, ble, &aseColProfile.reserved)
	if aseColProfile.Type == 2 {
		binary.Write(w, ble, &aseColProfile.ICCProfileDatLen)
		binary.Write(w, ble, &aseColProfile.ICCProfileDat)
	}
}

func (aseExtFile *AsepriteExternalFilesChunk2008) Decode(r io.Reader) {
	binary.Read(r, ble, &aseExtFile.NumEntries)
	binary.Read(r, ble, &aseExtFile.reserved)
	// for each entry
	aseExtFile.ExternalFile = make([]AsepriteExternalFilesChunk2008Entry, aseExtFile.NumEntries)
	for _, file := range aseExtFile.ExternalFile {
		binary.Read(r, ble, &file.EntryID)
		binary.Read(r, ble, &file.reserved)
		file.ExternalFilename = DecodeAseString(r)
	}
}

func (aseExtFile *AsepriteExternalFilesChunk2008) Encode(w io.Writer) {
	binary.Write(w, ble, &aseExtFile.NumEntries)
	binary.Write(w, ble, &aseExtFile.reserved)
	// for each entry
	for _, file := range aseExtFile.ExternalFile {
		binary.Write(w, ble, &file.EntryID)
		binary.Write(w, ble, &file.reserved)
		EncodeAseString(w, file.ExternalFilename)
	}
}

func (aseMask *AsepriteMaskChunk2016) Decode(r io.Reader) {
	binary.Read(r, ble, &aseMask.X)
	binary.Read(r, ble, &aseMask.Y)
	binary.Read(r, ble, &aseMask.Width)
	binary.Read(r, ble, &aseMask.Height)
	binary.Read(r, ble, &aseMask.future)
	aseMask.MaskName = DecodeAseString(r)
	aseMask.BitMapData = make([]byte,
		(aseMask.Height * ((aseMask.Width + 7) / 8)))
	binary.Read(r, ble, &aseMask.BitMapData)
}

func (aseMask *AsepriteMaskChunk2016) Encode(w io.Writer) {
	binary.Write(w, ble, &aseMask.X)
	binary.Write(w, ble, &aseMask.Y)
	binary.Write(w, ble, &aseMask.Width)
	binary.Write(w, ble, &aseMask.Height)
	binary.Write(w, ble, &aseMask.future)
	EncodeAseString(w, aseMask.MaskName)
	binary.Write(w, ble, &aseMask.BitMapData)
}

func (aseTags *AsepriteTagsChunk2018) Decode(r io.Reader) {
	binary.Read(r, ble, &aseTags.NumTags)
	binary.Read(r, ble, &aseTags.reserved1)
	aseTags.Tags = make([]AsepriteTagsChunk2018Tag, aseTags.NumTags)
	for x := 0; x < int(aseTags.NumTags); x += 1 {
		aseTags.Tags[x].Decode(r)
	}
}

func (aseTag *AsepriteTagsChunk2018Tag) Decode(r io.Reader) {
	binary.Read(r, ble, &aseTag.FromFrame)
	binary.Read(r, ble, &aseTag.ToFrame)
	binary.Read(r, ble, &aseTag.LoopAnimDirection)
	binary.Read(r, ble, &aseTag.reserved2)
	binary.Read(r, ble, &aseTag.TagColor)
	binary.Read(r, ble, &aseTag.ExtraByte)
	aseTag.TagName = DecodeAseString(r)
}

func (aseTags AsepriteTagsChunk2018) Encode(w io.Writer) {
	binary.Write(w, ble, &aseTags.NumTags)
	binary.Write(w, ble, &aseTags.reserved1)
	for _, tag := range aseTags.Tags {
		tag.Encode(w)
	}
}

func (aseTag AsepriteTagsChunk2018Tag) Encode(w io.Writer) {
	binary.Write(w, ble, &aseTag.FromFrame)
	binary.Write(w, ble, &aseTag.ToFrame)
	binary.Write(w, ble, &aseTag.LoopAnimDirection)
	binary.Write(w, ble, &aseTag.reserved2)
	binary.Write(w, ble, &aseTag.TagColor)
	binary.Write(w, ble, &aseTag.ExtraByte)
	EncodeAseString(w, aseTag.TagName)
}

func (asePaletteChunk *AsepritePaletteChunk2019) Decode(r io.Reader) {
	binary.Read(r, ble, &asePaletteChunk.PaletteSize)
	binary.Read(r, ble, &asePaletteChunk.FirstColIndexToChange)
	binary.Read(r, ble, &asePaletteChunk.LastColIndexToChange)
	binary.Read(r, ble, &asePaletteChunk.reserved)
	asePaletteChunk.PaletteEntries =
		make([]AsepritePaletteChunk2019Entry, asePaletteChunk.PaletteSize)
	for x := 0; x < len(asePaletteChunk.PaletteEntries); x += 1 {
		asePaletteChunk.PaletteEntries[x].Decode(r)
	}
}

func (asePaletteEntry *AsepritePaletteChunk2019Entry) Decode(r io.Reader) {
	binary.Read(r, ble, &asePaletteEntry.EntryFlags)
	binary.Read(r, ble, &asePaletteEntry.R)
	binary.Read(r, ble, &asePaletteEntry.G)
	binary.Read(r, ble, &asePaletteEntry.B)
	binary.Read(r, ble, &asePaletteEntry.A)
	if asePaletteEntry.EntryFlags&0x01 == 1 {
		asePaletteEntry.ColorName = DecodeAseString(r)
	}
}

func (asePaletteChunk AsepritePaletteChunk2019) Encode(w io.Writer) {
	binary.Write(w, ble, &asePaletteChunk.PaletteSize)
	binary.Write(w, ble, &asePaletteChunk.FirstColIndexToChange)
	binary.Write(w, ble, &asePaletteChunk.LastColIndexToChange)
	binary.Write(w, ble, &asePaletteChunk.reserved)
	for _, paletteEntry := range asePaletteChunk.PaletteEntries {
		paletteEntry.Encode(w)
	}
}

func (asePaletteEntry AsepritePaletteChunk2019Entry) Encode(w io.Writer) {
	binary.Write(w, ble, &asePaletteEntry.EntryFlags)
	binary.Write(w, ble, &asePaletteEntry.R)
	binary.Write(w, ble, &asePaletteEntry.G)
	binary.Write(w, ble, &asePaletteEntry.B)
	binary.Write(w, ble, &asePaletteEntry.A)
	EncodeAseString(w, asePaletteEntry.ColorName)
}

func (aseUserDat *AsepriteUserDataChunk2020) Decode(r io.Reader) {
	binary.Read(r, ble, &aseUserDat.Flags)
	if aseUserDat.Flags&0x00000001 == 1 {
		aseUserDat.Text = DecodeAseString(r)
	}
	if aseUserDat.Flags&0x00000002 == 2 {
		binary.Read(r, ble, &aseUserDat.R)
		binary.Read(r, ble, &aseUserDat.G)
		binary.Read(r, ble, &aseUserDat.B)
		binary.Read(r, ble, &aseUserDat.A)
	}
}

func (aseUserDat AsepriteUserDataChunk2020) Encode(w io.Writer) {
	binary.Write(w, ble, &aseUserDat.Flags)
	if aseUserDat.Flags&0x00000001 == 1 {
		EncodeAseString(w, aseUserDat.Text)
	}
	if aseUserDat.Flags&0x00000002 == 2 {
		binary.Write(w, ble, &aseUserDat.R)
		binary.Write(w, ble, &aseUserDat.G)
		binary.Write(w, ble, &aseUserDat.B)
		binary.Write(w, ble, &aseUserDat.A)
	}
}

func (aseSlice *AsepriteSliceChunk2022) Decode(r io.Reader) {
	binary.Read(r, ble, &aseSlice.NumSliceKeys)
	binary.Read(r, ble, &aseSlice.Flags)
	binary.Read(r, ble, &aseSlice.reserved)
	aseSlice.Name = DecodeAseString(r)
	aseSlice.SliceKeysData =
		make([]AsepriteSliceChunk2022Data, aseSlice.NumSliceKeys)
	for i, slice := range aseSlice.SliceKeysData {
		slice.parentChunk = aseSlice
		slice.Decode(r)
		aseSlice.SliceKeysData[i] = slice
	}
}

func (aseSliceDat *AsepriteSliceChunk2022Data) Decode(r io.Reader) {
	binary.Read(r, ble, &aseSliceDat.FrameNumber)
	binary.Read(r, ble, &aseSliceDat.SliceXOriginCoords)
	binary.Read(r, ble, &aseSliceDat.SliceYOriginCoords)
	binary.Read(r, ble, &aseSliceDat.SliceWidth)
	binary.Read(r, ble, &aseSliceDat.SliceHeight)
	if aseSliceDat.parentChunk.Flags&0x00000001 == 1 {
		binary.Read(r, ble, &aseSliceDat.CenterX)
		binary.Read(r, ble, &aseSliceDat.CenterY)
		binary.Read(r, ble, &aseSliceDat.CenterWidth)
		binary.Read(r, ble, &aseSliceDat.CenterHeight)
	}
	if aseSliceDat.parentChunk.Flags&0x00000002 == 2 {
		binary.Read(r, ble, &aseSliceDat.PivotX)
		binary.Read(r, ble, &aseSliceDat.PivotY)
	}
}

func (aseSlice AsepriteSliceChunk2022) Encode(w io.Writer) {
	binary.Write(w, ble, &aseSlice.NumSliceKeys)
	binary.Write(w, ble, &aseSlice.Flags)
	binary.Write(w, ble, &aseSlice.reserved)
	EncodeAseString(w, aseSlice.Name)
	for _, slice := range aseSlice.SliceKeysData {
		slice.parentChunk = &aseSlice
		slice.Encode(w)
	}
}

func (aseSliceDat AsepriteSliceChunk2022Data) Encode(w io.Writer) {
	binary.Write(w, ble, &aseSliceDat.FrameNumber)
	binary.Write(w, ble, &aseSliceDat.SliceXOriginCoords)
	binary.Write(w, ble, &aseSliceDat.SliceYOriginCoords)
	binary.Write(w, ble, &aseSliceDat.SliceWidth)
	binary.Write(w, ble, &aseSliceDat.SliceHeight)
	if aseSliceDat.parentChunk.Flags&0x00000001 == 1 {
		binary.Write(w, ble, &aseSliceDat.CenterX)
		binary.Write(w, ble, &aseSliceDat.CenterY)
		binary.Write(w, ble, &aseSliceDat.CenterWidth)
		binary.Write(w, ble, &aseSliceDat.CenterHeight)
	}
	if aseSliceDat.parentChunk.Flags&0x00000002 == 2 {
		binary.Write(w, ble, &aseSliceDat.PivotX)
		binary.Write(w, ble, &aseSliceDat.PivotY)
	}
}

func (aseTileset *AsepriteTilesetChunk2023) Decode(r io.Reader) {
	binary.Read(r, ble, &aseTileset.TilesetID)
	binary.Read(r, ble, &aseTileset.Flags)
	binary.Read(r, ble, &aseTileset.NumTiles)
	binary.Read(r, ble, &aseTileset.TileWidth)
	binary.Read(r, ble, &aseTileset.TileHeight)
	binary.Read(r, ble, &aseTileset.BaseIndex)
	binary.Read(r, ble, &aseTileset.reserved)
	aseTileset.Name = DecodeAseString(r)
	if aseTileset.Flags&0x00000001 == 1 {
		binary.Read(r, ble, &aseTileset.ExternalFileID)
		binary.Read(r, ble, &aseTileset.TilesetIDInExternalFile)
	}
	if aseTileset.Flags&0x00000002 == 2 {
		binary.Read(r, ble, &aseTileset.CompressedDatLen)
		aseTileset.CompressedTilesetImg = make([]byte, aseTileset.CompressedDatLen)
		binary.Read(r, ble, &aseTileset.CompressedTilesetImg)
	}
}
