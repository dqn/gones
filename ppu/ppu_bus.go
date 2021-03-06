package ppu

// https://qiita.com/bokuweb/items/1575337bef44ae82f4d3#%E3%83%A1%E3%83%A2%E3%83%AA%E3%83%9E%E3%83%83%E3%83%97-1

// アドレス	       サイズ   用途
// 0x0000～0x0FFF	0x1000	パターンテーブル 0
// 0x1000～0x1FFF	0x1000	パターンテーブル 1
// 0x2000～0x23BF	0x03C0	ネームテーブル 0
// 0x23C0～0x23FF	0x0040	属性テーブル 0
// 0x2400～0x27BF	0x03C0	ネームテーブル 1
// 0x27C0～0x27FF	0x0040	属性テーブル 1
// 0x2800～0x2BBF	0x03C0	ネームテーブル 2
// 0x2BC0～0x2BFF	0x0040	属性テーブル 2
// 0x2C00～0x2FBF	0x03C0	ネームテーブル 3
// 0x2FC0～0x2FFF	0x0040	属性テーブル 3
// 0x3000～0x3EFF	-	      0x2000~0x2EFF のミラー
// 0x3F00～0x3F0F	0x0010	バックグラウンドパレット
// 0x3F10～0x3F1F	0x0010	スプライトパレット
// 0x3F20～0x3FFF	-	      0x3F00~0x3F1F のミラー

type vram [0x4000]uint8

type PPUBus struct {
	vram         *vram
	characterROM []uint8
}

func NewBus(characterROM []uint8) *PPUBus {
	return &PPUBus{&vram{}, characterROM}
}

func (b *PPUBus) Read(addr uint16) uint8 {
	switch {
	case addr < 0x2000:
		return b.characterROM[addr]
	case addr < 0x3F00:
		return b.vram[addr]
	case addr < 0x4000:
		return b.vram[addr]
	}

	println("!!!", addr)
	panic(1)
}

func (b *PPUBus) Write(addr uint16, data uint8) {
	// fmt.Printf("!!! %x 0x%x\n", addr, data)
	switch {
	case addr < 0x2000:
	case addr < 0x4000:
		b.vram[addr] = data
	}
}
