package asefile

// Some color definitions

// Just used here by certain packets
type AsepriteRGB24 struct {
	R, G, B byte
}

/**
 * ASE files use Intel (little-endian) byte order.
 *
 * BYTE: An 8-bit unsigned integer value
 * WORD: A 16-bit unsigned integer value
 * SHORT: A 16-bit signed integer value
 * DWORD: A 32-bit unsigned integer value
 * LONG: A 32-bit signed integer value
 * FIXED: A 32-bit fixed point (16.16) value
 * BYTE[n]: "n" bytes.
 * STRING:
 *   WORD: string length (number of bytes)
 *   BYTE[length]: characters (in UTF-8) The '\0' character is not included.
 * PIXEL: One pixel, depending on the image pixel format:
 *   RGBA: BYTE[4], each pixel have 4 bytes in this order Red, Green, Blue, Alpha.
 *   Grayscale: BYTE[2], each pixel have 2 bytes in the order Value, Alpha.
 *   Indexed: BYTE, each pixel uses 1 byte (the index).
 * TILE: Tilemaps: Each tile can be a 8-bit (BYTE), 16-bit (WORD), or 32-bit (DWORD) value and there are masks
 * related to the meaning of each bit.
 *
 * File Header:
 * DWORD     File size
 * WORD      Magic number (0xA5E0)
 * WORD      Frames
 * WORD      Width in pixels
 * WORD      Height in pixels
 * WORD      Color depth (bits per pixel)
 *           32 bpp = RGBA
 *           16 bpp = Grayscale
 *           8 bpp = Indexed
 * DWORD     Flags:
 *           1 = Layer opacity has valid value
 * WORD      Speed (milliseconds between frame, like in FLC files)
 *           DEPRECATED: You should use the frame duration field
 *           from each frame header
 * DWORD     Set be 0
 * DWORD     Set be 0
 * BYTE      Palette entry (index) which represent transparent color
 *           in all non-background layers (only for Indexed sprites).
 * BYTE[3]   Ignore these bytes
 * WORD      Number of colors (0 means 256 for old sprites)
 * BYTE      Pixel width (pixel ratio is "pixel width/pixel height").
 *           If this or pixel height field is zero, pixel ratio is 1:1
 * BYTE      Pixel height
 * SHORT     X position of the grid
 * SHORT     Y position of the grid
 * WORD      Grid width (zero if there is no grid, grid size
 *           is 16x16 on Aseprite by default)
 * WORD      Grid height (zero if there is no grid)
 * BYTE[84]  For future (set to zero)
 */

type AsepriteHeader struct {
	FileSize       uint32
	MagicNumber    uint16 // A5E0
	Frames         uint16
	WidthInPixels  uint16
	HeightInPixels uint16
	ColorDepth     uint16
	Flags          uint32
	Speed          uint16 // deprecated, use frame duration from frame header
	// These are 0
	ignore1, ignore2 uint32
	PaletteEntry     byte // represents transparent color in all non-background layers (only for indexed sprites)
	// format defines these 3 bytes as ignored
	ignore3        [3]byte
	NumberOfColors uint16
	// (pixel ratio is "pxel width/pixel height")
	PixelWidth, PixelHeight byte
	XPositionOfGrid         int16
	YPositionOfGrid         int16
	GridWidth, GridHeight   uint16
	// Future reserved
	reserved [84]byte
}

/**
 * DWORD     Bytes in this frame
 * WORD      Magic number (always 0xF1FA)
 * WORD      Old field which specifies the number of "chunks"
 *           in this frame. If this value is 0xFFFF, we might
 *           have more chunks to read in this frame
 *           (so we have to use the new field)
 * WORD      Frame duration (in milliseconds)
 * BYTE[2]   For future (set to zero)
 * DWORD     New field which specifies the number of "chunks"
 *           in this frame (if this is 0, use the old field)
 */

type AsepriteFrame struct {
	parentHeader              *AsepriteHeader
	BytesThisFrame            uint32
	MagicNumber               uint16 // F1FA
	ChunksThisFrame           uint16 // If this value is FFFF there "may" be more chunks to read
	FrameDurationMilliseconds uint16
	reserved                  [2]byte
	ChunksThisFrameExt        uint32 // New field which specifies num chunks. If 0, use old field
	OldPalettes0004           []AsepriteOldPaletteChunk0004
	OldPalettes0011           []AsepritePaletteChunk0011
	Layers                    []AsepriteLayerChunk2004
	Cels                      []AsepriteCelChunk2005
	ColorProfiles             []AsepriteColorProfileChunk2007
	Tags                      AsepriteTagsChunk2018
	Palettes                  []AsepritePaletteChunk2019
}

/**
 * Then each chunk format is:
 * DWORD       Chunk size
 * WORD        Chunk type
 * BYTE[]      Chunk data
 *
 * -----------------
 *
 * Ignore this chunk if you find the new palette chunk (0x2019) Aseprite v1.1 saves both chunks 0x0004 and 0x2019 just for backward compatibility.
 * WORD        Number of packets
 * + For each packet
 *  BYTE      Number of palette entries to skip from the last packet (start from 0)
 *  BYTE      Number of colors in the packet (0 means 256)
 *  + For each color in the packet
 *    BYTE    Red (0-255)
 *    BYTE    Green (0-255)
 *    BYTE    Blue (0-255)
 */
type AsepriteOldPaletteChunk0004 struct {
	NumberOfPackets uint16
	// + for each packet
	Packets []AsepriteOldPaletteChunk0004Packet
}

type AsepriteOldPaletteChunk0004Packet struct {
	NumPalletteEntriesToSkip byte // from thee last packet (start from 0)
	NumColorsInThePacket     byte // 0 means 256
	//   + for each color in the packet
	Colors []AsepriteRGB24
}

/**
 * Old palette chunk (0x0011)
 * Ignore this chunk if you find the new palette chunk (0x2019)
 * WORD        Number of packets
 * + For each packet
 * BYTE      Number of palette entries to skip from the last packet (start from 0)
 * BYTE      Number of colors in the packet (0 means 256)
 * + For each color in the packet
 *   BYTE    Red (0-63)
 *   BYTE    Green (0-63)
 *   BYTE    Blue (0-63)
 */

type AsepritePaletteChunk0011 struct {
	NumberOfPackets uint16
	// + for each packet
	Packets []AsepritePaletteChunk0011Packet
}

type AsepritePaletteChunk0011Packet struct {
	NumPalletteEntriesToSkip byte // start from 0
	NumColorsInThePacket     byte // 0 means 256
	// + for each color in the packet
	Colors []AsepriteRGB24 // but using colors in the range 0-63
}

/**
 * Layer Chunk (0x2004)
 * In the first frame should be a set of layer chunks to determine the entire layers layout:
 * WORD  Flags:
 *        1 = Visible
 *        2 = Editable
 *        4 = Lock movement
 *        8 = Background
 *       16 = Prefer linked cels
 *       32 = The layer group should be displayed collapsed
 *       64 = The layer is a reference layer
 * WORD  Layer type
 *        0 = Normal (image) layer
 *        1 = Group
 *        2 = Tilemap
 * WORD  Layer child level (see NOTE.1)
 * WORD  Default layer width in pixels (ignored)
 * WORD  Default layer height in pixels (ignored)
 * WORD  Blend mode (always 0 for layer set)
 *        Normal       = 0   Multiply     = 1
 *        Screen       = 2   Overlay      = 3
 *        Darken       = 4   Lighten      = 5
 *        Color Dodge  = 6   Color Burn   = 7
 *        Hard Light   = 8   Soft Light   = 9
 *        Difference   = 10  Exclusion    = 11
 *        Hue          = 12  Saturation   = 13
 *        Color        = 14  Luminosity   = 15
 *        Addition     = 16  Subtract     = 17
 *        Divide       = 18
 * BYTE  Opacity
 *        Note: valid only if file header flags field has bit 1 set
 * BYTE[3] For future (set to zero)
 * STRING  Layer name
 *  + If layer type = 2
 * DWORD   Tileset index
 */

type AsepriteLayerChunk2004 struct {
	Flags                uint16
	LayerType            uint16
	LayerChildLevel      uint16
	DefLayerWidthPixels  uint16 // (ignored)
	DefLayerHeightPixels uint16 // (ignored)
	BlendMode            uint16
	Opacity              byte // only valid if file headre flag field has bit 1 set
	forFuture            [3]byte
	LayerName            string
	// + if layer type = 2
	TilesetIndex uint32
	UserData     AsepriteUserDataChunk2020
}

func (layer *AsepriteLayerChunk2004) AddUserData(userData AsepriteUserDataChunk2020) {
	layer.UserData = userData
}

/**
 * Cel Chunk (0x2005)
 * This chunk determine where to put a cel in the specified layer/frame.
 *
 * WORD        Layer index (see NOTE.2)
 * SHORT       X position
 * SHORT       Y position
 * BYTE        Opacity level
 * WORD        Cel Type
 *           0 - Raw Image Data (unused, compressed image is preferred)
 *           1 - Linked Cel
 *           2 - Compressed Image
 *           3 - Compressed Tilemap
 * BYTE[7]     For future (set to zero)
 * + For cel type = 0 (Raw Image Data)
 *  WORD      Width in pixels
 *  WORD      Height in pixels
 *  PIXEL[]   Raw pixel data: row by row from top to bottom,
 *           for each scanline read pixels from left to right.
 * + For cel type = 1 (Linked Cel)
 *  WORD      Frame position to link with
 * + For cel type = 2 (Compressed Image)
 *  WORD      Width in pixels
 *  WORD      Height in pixels
 *  BYTE[]    "Raw Cel" data compressed with ZLIB method (see NOTE.3)
 * + For cel type = 3 (Compressed Tilemap)
 *  WORD      Width in number of tiles
 *  WORD      Height in number of tiles
 *  WORD      Bits per tile (at the moment it's always 32-bit per tile)
 *  DWORD     Bitmask for tile ID (e.g. 0x1fffffff for 32-bit tiles)
 *  DWORD     Bitmask for X flip
 *  DWORD     Bitmask for Y flip
 *  DWORD     Bitmask for 90CW rotation
 *  BYTE[10]  Reserved
 *  TILE[]    Row by row, from top to bottom tile by tile
 *            compressed with ZLIB method (see NOTE.3)
 */

type AsepriteCelChunk2005 struct {
	parentHeader *AsepriteHeader
	chunkSize    uint32
	LayerIndex   uint16
	X, Y         int16
	OpacityLevel byte
	CelType      uint16
	future       [7]byte
	// + For cel type = 0 (Raw Image Data)
	WidthInPix, HeightInPix uint16
	RawPixData              []byte
	// + For cel type = 1 (Linked Cel)
	FramePosToLinkWith uint16
	// + For cel type = 2 (Compressed Image)
	// WORD      Width in pixels
	// WORD      Height in pixels
	RawCelData []byte // "Raw Cel" data decompressed from ZLIB (see NOTE.3)
	// + For cel type = 3 (Compressed Tilemap)
	WidthInTiles, HeightInTiles uint16
	BitsPerTile                 uint16 // (at the moment it's always 32-bit per tile)
	BitMaskForTileID            uint32 // (e.g. 0x1fffffff for 32-bit tiles)
	BitMaskForXFlip             uint32
	BitMaskForYFlip             uint32
	BitMaskFor90CWRot           uint32
	reserved                    [10]byte
	Tiles                       []byte // zlib data (see NOTE.3)
	Extra                       *AsepriteCelExtraChunk2006
}

/**
 * Cel Extra Chunk (0x2006)
 * Adds extra information to the latest read cel.
 *
 * DWORD     Flags (set to zero)
 *            1 = Precise bounds are set
 * FIXED     Precise X position
 * FIXED     Precise Y position
 * FIXED     Width of the cel in the sprite (scaled in real-time)
 * FIXED     Height of the cel in the sprite
 * BYTE[16]  For future use (set to zero)
 */
type AsepriteCelExtraChunk2006 struct {
	Flags              uint32
	PreciseX, PreciseY uint32
	WidthCelInSprite   uint32
	HeightCelInSprite  uint32
	futureUse          [16]byte
}

/**
 * Color Profile Chunk (0x2007)
 * Color profile for RGB or grayscale values.
 *
 * WORD        Type
 *              0 - no color profile (as in old .aseprite files)
 *              1 - use sRGB
 *              2 - use the embedded ICC profile
 * WORD        Flags
 *              1 - use special fixed gamma
 * FIXED       Fixed gamma (1.0 = linear)
 *             Note: The gamma in sRGB is 2.2 in overall but it doesn't use
 *             this fixed gamma, because sRGB uses different gamma sections
 *             (linear and non-linear). If sRGB is specified with a fixed
 *             gamma = 1.0, it means that this is Linear sRGB.
 * BYTE[8]     Reserved (set to zero)
 * + If type = ICC:
 * DWORD     ICC profile data length
 * BYTE[]    ICC profile data. More info: http://www.color.org/ICC1V42.pdf
 */

type AsepriteColorProfileChunk2007 struct {
	Type       uint16
	Flags      uint16
	FixedGamma uint32
	reserved   [8]byte
	// + If type = ICC:
	ICCProfileDatLen uint32
	ICCProfileDat    []byte
}

/**
 * External Files Chunk (0x2008)
 * A list of external files linked with this file. It might be used to reference external palettes or tilesets.
 *
 * DWORD       Number of entries
 * BYTE[8]     Reserved (set to zero)
 * + For each entry
 *  DWORD     Entry ID (this ID is referenced by tilesets or palettes)
 *  BYTE[8]   Reserved (set to zero)
 *  STRING    External file name
 */

type AsepriteExternalFilesChunk2008 struct {
	NumEntries uint32
	reserved   [8]byte
	// + for each entry
	ExternalFile []AsepriteExternalFilesChunk2008Entry
}

type AsepriteExternalFilesChunk2008Entry struct {
	EntryID          uint32
	reserved         [8]byte
	ExternalFilename string
}

/**
 * Mask Chunk (0x2016) DEPRECATED
 *
 * SHORT   X position
 * SHORT   Y position
 * WORD    Width
 * WORD    Height
 * BYTE[8] For future (set to zero)
 * STRING  Mask name
 * BYTE[]  Bit map data (size = height*((width+7)/8))
 *         Each byte contains 8 pixels (the leftmost pixels are
 *         packed into the high order bits)
 */

type AsepriteMaskChunk2016 struct {
	X, Y          int16
	Width, Height uint16
	future        [8]byte
	MaskName      string
	BitMapData    []byte
}

/**
 * Path Chunk (0x2017)
 * Never used
 */

type AsepritePathChunk2017 struct{}

/**
 * Tags Chunk (0x2018)
 * After the tags chunk, you can write one user data chunk for each tag. E.g. if there are 10 tags, you can then write 10 user data chunks one for each tag.
 *
 * WORD        Number of tags
 * BYTE[8]     For future (set to zero)
 * + For each tag
 *  WORD      From frame
 *  WORD      To frame
 *  BYTE      Loop animation direction
 *              0 = Forward
 *              1 = Reverse
 *              2 = Ping-pong
 *  BYTE[8]   For future (set to zero)
 *  BYTE[3]   RGB values of the tag color
 *              Deprecated, used only for backward compatibility with Aseprite v1.2.x
 *              The color of the tag is the one in the user data field following
 *              the tags chunk
 *  BYTE      Extra byte (zero)
 *  STRING    Tag name
 */
type AsepriteTagsChunk2018 struct {
	NumTags   uint16
	reserved1 [8]byte
	Tags      []AsepriteTagsChunk2018Tag
	UserData  []AsepriteUserDataChunk2020
}

func (tagsChunk *AsepriteTagsChunk2018) AddUserData(userDat AsepriteUserDataChunk2020) {
	tagsChunk.UserData = append(tagsChunk.UserData, userDat)
}

type AsepriteTagsChunk2018Tag struct {
	FromFrame, ToFrame uint16
	LoopAnimDirection  byte
	reserved2          [8]byte
	TagColor           [3]byte // deprecated
	ExtraByte          byte    // (zero)
	TagName            string
}

/**
 * Palette Chunk (0x2019)
 *
 * DWORD       New palette size (total number of entries)
 * DWORD       First color index to change
 * DWORD       Last color index to change
 * BYTE[8]     For future (set to zero)
 * + For each palette entry in [from,to] range (to-from+1 entries)
 *   WORD      Entry flags:
 *               1 = Has name
 *   BYTE      Red (0-255)
 *   BYTE      Green (0-255)
 *   BYTE      Blue (0-255)
 *   BYTE      Alpha (0-255)
 *   + If has name bit in entry flags
 *     STRING  Color name
 */

type AsepritePaletteChunk2019 struct {
	PaletteSize           uint32
	FirstColIndexToChange uint32
	LastColIndexToChange  uint32
	reserved              [8]byte
	// + For each palette entry in [from,to] range (to-from+1 entries)
	PaletteEntries []AsepritePaletteChunk2019Entry
}

type AsepritePaletteChunk2019Entry struct {
	EntryFlags uint16
	R, G, B, A byte
	// + If has name bit in entry flags
	ColorName string
}

/**
 * User Data Chunk (0x2020)
 * Insert this user data in the last read chunk. E.g. If we've read a layer, this user data belongs to that layer, if we've read a cel, it belongs to that cel, etc. There are some special cases: After a Tags chunk, there will be several user data fields, one for each tag, you should associate the user data in the same order as the tags are in the Tags chunk. In version 1.3 a sprite has associated user data, to consider this case there is an User Data Chunk at the first frame after the Palette Chunk.
 *
 * DWORD  Flags
 *         1 = Has text  2 = Has color
 *  + If flags have bit 1
 *    STRING    Text
 *  + If flags have bit 2
 *    BYTE  Color Red (0-255)
 *    BYTE  Color Green (0-255)
 *    BYTE  Color Blue (0-255)
 *    BYTE  Color Alpha (0-255)
 */

type AsepriteUserDataChunk2020 struct {
	Flags uint32
	// + If flags have bit 1
	Text string
	// + If flags have bit 2
	R, G, B, A byte
}

/**
 * Slice Chunk (0x2022)
 *
 * DWORD       Number of "slice keys"
 * DWORD       Flags
 *              1 = It's a 9-patches slice
 *              2 = Has pivot information
 * DWORD       Reserved
 * STRING      Name
 * + For each slice key
 *   DWORD     Frame number (this slice is valid from
 *             this frame to the end of the animation)
 *   LONG      Slice X origin coordinate in the sprite
 *   LONG      Slice Y origin coordinate in the sprite
 *   DWORD     Slice width (can be 0 if this slice hidden in the
 *             animation from the given frame)
 *   DWORD     Slice height
 *   + If flags have bit 1
 *     LONG    Center X position (relative to slice bounds)
 *     LONG    Center Y position
 *     DWORD   Center width
 *     DWORD   Center height
 *   + If flags have bit 2
 *     LONG    Pivot X position (relative to the slice origin)
 *     LONG    Pivot Y position (relative to the slice origin)
 */

type AsepriteSliceChunk2022 struct {
	NumSliceKeys uint32
	Flags        uint32
	reserved     uint32
	Name         string
	// + For each slice key
	SliceKeysData []AsepriteSliceChunk2022Data
}

type AsepriteSliceChunk2022Data struct {
	parentChunk             *AsepriteSliceChunk2022
	FrameNumber             uint32
	SliceXOriginCoords      int32
	SliceYOriginCoords      int32
	SliceWidth, SliceHeight uint32
	// + If flags have bit 1
	CenterX, CenterY          int32
	CenterWidth, CenterHeight uint32
	// + If flags have bit 2
	PivotX, PivotY int32
}

/**
 * Tileset Chunk (0x2023)
 *
 * DWORD    Tileset ID
 * DWORD    Tileset flags
 *             1 - Include link to external file
 *             2 - Include tiles inside this file
 *             4 - Tilemaps using this tileset use tile ID=0 as empty tile
 *              (this is the new format). In rare cases this bit is off,
 *               and the empty tile will be equal to 0xffffffff (used in
 *               internal versions of Aseprite)
 * DWORD    Number of tiles
 * WORD     Tile Width
 * WORD     Tile Height
 * SHORT    Base Index: Number to show in the screen from the tile with
 *           index 1 and so on (by default this is field is 1, so the data
 *           that is displayed is equivalent to the data in memory). But it
 *           can be 0 to display zero-based indexing (this field isn't used
 *           for the representation of the data in the file, it's just for
 *           UI purposes).
 * BYTE[14] Reserved
 * STRING   Name of the tileset
 * + If flag 1 is set
 *   DWORD  ID of the external file. This ID is one entry
 *            of the the External Files Chunk.
 *   DWORD  Tileset ID in the external file
 * + If flag 2 is set
 *   DWORD   Compressed data length
 *   PIXEL[] Compressed Tileset image (see NOTE.3):
 *               (Tile Width) x (Tile Height x Number of Tiles)
 */

type AsepriteTilesetChunk2023 struct {
	TilesetID             uint32
	Flags                 uint32
	NumTiles              uint32
	TileWidth, TileHeight uint16
	BaseIndex             int16
	reserved              [14]byte
	Name                  string
	// + If flag 1 is set
	ExternalFileID          uint32
	TilesetIDInExternalFile uint32
	// + If flag 2 is set
	CompressedDatLen     uint32
	CompressedTilesetImg []byte
}

/**
 * Notes
 * NOTE.1
 * The child level is used to show the relationship of this layer with the last one read, for example:
 *
 * Layer name and hierarchy      Child Level
 * -----------------------------------------------
 * - Background                  0
 *   `- Layer1                   1
 * - Foreground                  0
 *   |- My set1                  1
 *   |  `- Layer2                2
 *   `- Layer3                   1
 *
 * NOTE.2
 * The layer index is a number to identify any layer in the sprite, for example:
 *
 * Layer name and hierarchy      Layer index
 * -----------------------------------------------
 * - Background                  0
 *   `- Layer1                   1
 * - Foreground                  2
 *   |- My set1                  3
 *   |  `- Layer2                4
 *   `- Layer3                   5
 *
 * NOTE.3
 * Details about the ZLIB and DEFLATE compression methods:
 * https://www.ietf.org/rfc/rfc1950
 * https://www.ietf.org/rfc/rfc1951
 * Some extra notes that might help you to decode the data: http://george.chiramattel.com/blog/2007/09/deflatestream-block-length-does-not-match.html
 */
