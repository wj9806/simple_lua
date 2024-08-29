package binchunk

//header使用
const (
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x54
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = 0x5678
	LUAC_NUM         = 370.5
)

//常量类型
const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_NUMBER    = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type binaryChunk struct {
	header                  //头部
	sizeUpvalues byte       //主函数upvalue数量
	mainFunc     *Prototype //主函数原型
}

type header struct {
	signature       [4]byte //签名
	version         byte    //版本
	format          byte    //格式号
	luacData        [6]byte //luac data
	cintSize        byte    //cint字节数
	sizetSize       byte    //size_t字节数
	instructionSize byte    //lua虚拟机指令字节数
	luaIntegerSize  byte    //lua整数字节数
	luaNumberSize   byte    //lua浮点数字节数
	luacInt         int64   //用于检查大小端
	luacNum         float64 //检测chunk使用的浮点数格式
}

//函数原型
type Prototype struct {
	Source          string //源文件名称
	LineDefined     uint32 //起止行号
	LastLineDefined uint32
	NumParams       byte          //固定参数个数
	IsVararg        byte          //是否有变长参数
	MaxStackSize    byte          //寄存器数量
	Code            []uint32      //指令表
	Constants       []interface{} //常量表
	Upvalues        []Upvalue     //闭包相关
	Protos          []*Prototype  //子函数原型表
	LineInfo        []uint32      //行号表
	LocVars         []LocVar      //局部变量表
	UpvalueNames    []string      //Upvalue名 列表
}

type Upvalue struct {
	Instack byte
	Idx     byte
}

type LocVar struct {
	VarName string
	StartPC uint32
	EndPC   uint32
}

func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()
	reader.readByte()
	return reader.readProto("")
}