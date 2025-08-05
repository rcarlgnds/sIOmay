package core

// ==========================IMPORTANT===============================
// Mapping and changing action to specific Opcode to use
// Schema:

//  operation  special key   data
//   1 byte      1 byte     5 byte
//    1-8          9-16     17-56

// Allocation
// More in excel
// 1.  Mouse Move = 0x0  | operation
// 	special key = (Optional) | special key
// 	Data: x[17-32], y[33-48]
// 2.
//===================================================================


type Opcode uint8
const (
    MouseMiddleClick Opcode = 0b00000001
    MouseScroll      Opcode = 0b00000010
    MouseClickRight  Opcode = 0b00000100
    MouseClickLeft   Opcode = 0b00001000
    MouseMove        Opcode = 0b00010000
)

type Bytedata struct {
    data [7]byte
}

func (d *Bytedata) MouseClickLeft() {
    d.data[0] |= byte(MouseClickLeft)
}

func (d *Bytedata) MouseClickRight() {
    d.data[0] |= byte(MouseClickRight)
}

func (d *Bytedata) MouseMove(x int16, y int16) {
    d.data[0] |= byte(MouseMove)

    // Encode X coordinate
    d.data[2] = byte(x >> 8)
    d.data[3] = byte(x & 0xFF)

    // Encode Y coordinate
    d.data[4] = byte(y >> 8)
    d.data[5] = byte(y & 0xFF)
}

// ========== Decoding (Read) Methods ==========

func (d *Bytedata) HasMouseClickLeft() bool {
    return d.data[0]&byte(MouseClickLeft) != 0
}

func (d *Bytedata) HasMouseClickRight() bool {
    return d.data[0]&byte(MouseClickRight) != 0
}

func (d *Bytedata) HasMouseMiddleClick() bool {
    return d.data[0]&byte(MouseMiddleClick) != 0
}

func (d *Bytedata) HasMouseScroll() bool {
    return d.data[0]&byte(MouseScroll) != 0
}

func (d *Bytedata) HasMouseMove() bool {
    return d.data[0]&byte(MouseMove) != 0
}

func (d *Bytedata) GetMousePosition() (x int16, y int16) {
    x = int16(d.data[2])<<8 | int16(d.data[3])
    y = int16(d.data[4])<<8 | int16(d.data[5])
    return
}

func (d *Bytedata) GetScrollRotation() int16 {
    return int16(d.data[2])<<8 | int16(d.data[3])
}

func (d *Bytedata) GetSpecialKey() byte {
    return d.data[1]
}

func (d *Bytedata) GetOpcode() Opcode {
    return Opcode(d.data[0])
}

func (d *Bytedata) Bytes() []byte {
    return d.data[:]
}

func (d *Bytedata) Clear() {
    for i := range d.data {
        d.data[i] = 0
    }
}

func NewBytedata() *Bytedata {
    return &Bytedata{}
}