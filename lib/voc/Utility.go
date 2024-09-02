package voc

const (
	fileHeader         string  = "Creative Voice File\u001A"
	standardHeaderSize uint16  = 0x1A
	baseVersion        uint16  = 0x010A
	versionCheckValue  uint16  = 0x1234
	rateBase           float32 = 1000000.0
)
