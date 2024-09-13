package binchunk

import (
	"encoding/binary"
	"math"
)

type reader struct {
	data []byte //存放被解析的二进制chunk数据
}

func (obj *reader) readByte() byte {
	b := obj.data[0]
	obj.data = obj.data[1:]
	return b
}

func (obj *reader) readBytes(n uint) []byte {
	bytes := obj.data[:n]
	obj.data = obj.data[n:]
	return bytes
}

func (obj *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(obj.data)
	obj.data = obj.data[4:]
	return i
}

func (obj *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(obj.data)
	obj.data = obj.data[8:]
	return i
}

func (obj *reader) readLuaInteger() int64 {
	return int64(obj.readUint64())
}

func (obj *reader) readLuaNumber() float64 {
	return math.Float64frombits(obj.readUint64())
}

/*
  - NULL字符串 0x00
    长度小于等于253字符串，先使用一个字节记录长度+1
    长度大于等于254字符串，第一个字节是0xFF,后面跟一个size_t记录长度+1
*/
func (obj *reader) readString() string {
	size := uint(obj.readByte())
	if size == 0 { //null字符串
		return ""
	}
	if size == 0xFF { //长字符串
		size = uint(obj.readUint64()) // size_t
	}
	bytes := obj.readBytes(size - 1)
	return string(bytes) // todo
}

func (obj *reader) checkHeader() {
	if string(obj.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk!")
	}
	if obj.readByte() != LUAC_VERSION {
		panic("version mismatch!")
	}
	if obj.readByte() != LUAC_FORMAT {
		panic("format mismatch!")
	}
	if string(obj.readBytes(6)) != LUAC_DATA {
		panic("corrupted!")
	}
	if obj.readByte() != CINT_SIZE {
		panic("int size mismatch!")
	}
	if obj.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch!")
	}
	if obj.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch!")
	}
	if obj.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch!")
	}
	if obj.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch!")
	}
	if obj.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch!")
	}
	if obj.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch!")
	}
}

func (obj *reader) readProto(parentSource string) *Prototype {
	source := obj.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefined:     obj.readUint32(),
		LastLineDefined: obj.readUint32(),
		NumParams:       obj.readByte(),
		IsVararg:        obj.readByte(),
		MaxStackSize:    obj.readByte(),
		Code:            obj.readCode(),
		Constants:       obj.readConstants(),
		Upvalues:        obj.readUpvalues(),
		Protos:          obj.readProtos(source),
		LineInfo:        obj.readLineInfo(),
		LocVars:         obj.readLocVars(),
		UpvalueNames:    obj.readUpvalueNames(),
	}
}

func (obj *reader) readCode() []uint32 {
	code := make([]uint32, obj.readUint32())
	for i := range code {
		code[i] = obj.readUint32()
	}
	return code
}

func (obj *reader) readConstants() []interface{} {
	constants := make([]interface{}, obj.readUint32())
	for i := range constants {
		constants[i] = obj.readConstant()
	}
	return constants
}

func (obj *reader) readConstant() interface{} {
	switch obj.readByte() {
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return obj.readByte() != 0
	case TAG_INTEGER:
		return obj.readLuaInteger()
	case TAG_NUMBER:
		return obj.readLuaNumber()
	case TAG_SHORT_STR, TAG_LONG_STR:
		return obj.readString()
	default:
		panic("corrupted!") // todo
	}
}

func (obj *reader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, obj.readUint32())
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: obj.readByte(),
			Idx:     obj.readByte(),
		}
	}
	return upvalues
}

func (obj *reader) readProtos(parentSource string) []*Prototype {
	protos := make([]*Prototype, obj.readUint32())
	for i := range protos {
		protos[i] = obj.readProto(parentSource)
	}
	return protos
}

func (obj *reader) readLineInfo() []uint32 {
	lineInfo := make([]uint32, obj.readUint32())
	for i := range lineInfo {
		lineInfo[i] = obj.readUint32()
	}
	return lineInfo
}

func (obj *reader) readLocVars() []LocVar {
	locVars := make([]LocVar, obj.readUint32())
	for i := range locVars {
		locVars[i] = LocVar{
			VarName: obj.readString(),
			StartPC: obj.readUint32(),
			EndPC:   obj.readUint32(),
		}
	}
	return locVars
}

func (obj *reader) readUpvalueNames() []string {
	names := make([]string, obj.readUint32())
	for i := range names {
		names[i] = obj.readString()
	}
	return names
}
