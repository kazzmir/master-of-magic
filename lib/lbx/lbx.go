package lbx

import (
    "fmt"
    "io"
    "bytes"
    "encoding/binary"
)

type LbxFile struct {
    Signature uint32
    Version uint16
    Data [][]byte
}

const LbxSignature = 0x0000fead

func readUint16(reader io.Reader) (uint16, error) {
    var value uint16
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readUint32(reader io.Reader) (uint32, error) {
    var value uint32
    err := binary.Read(reader, binary.LittleEndian, &value)
    return value, err
}

func readInt32(reader io.Reader) (int32, error) {
    v, err := readUint32(reader)
    return int32(v), err
}

func ReadLbx(reader io.ReadSeeker) (LbxFile, error) {
    numFiles, err := readUint16(reader)
    if err != nil {
        return LbxFile{}, err
    }

    signature, err := readUint32(reader)
    if err != nil {
        return LbxFile{}, err
    }

    if signature != LbxSignature {
        return LbxFile{}, fmt.Errorf("Invalid lbx signature, was 0x%x but expected 0x%x\n", signature, LbxSignature)
    }

    version, err := readUint16(reader)
    if err != nil {
        return LbxFile{}, err
    }

    // fmt.Printf("Version: %v\n", version)

    var offsets []uint32

    for i := 0; i < int(numFiles); i++ {
        offset, err := readUint32(reader)
        if err != nil {
            return LbxFile{}, err
        }

        // fmt.Printf("Offset %v: 0x%x\n", i, offset)

        offsets = append(offsets, offset)
    }

    reader.Seek(0, io.SeekEnd)
    lastByte, _ := reader.Seek(0, io.SeekCurrent)

    var lbx LbxFile

    lbx.Signature = signature
    lbx.Version = version

    for i, offset := range offsets {
        reader.Seek(int64(offset), io.SeekStart)
        end := uint32(0)
        if i < len(offsets) - 1 {
            end = offsets[i+1]
        } else {
            end = uint32(lastByte)
        }

        limitedReader := io.LimitReader(reader, int64(end - offset))
        var buffer bytes.Buffer
        io.Copy(&buffer, limitedReader)

        // fmt.Printf("File %v: %v bytes\n", i, buffer.Len())
        lbx.Data = append(lbx.Data, buffer.Bytes())
    }

    return lbx, nil
}
