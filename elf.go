package elfwriter

import (
    "bytes"
    "encoding/binary"
    "debug/elf"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

var fileHeaderSize, programHeaderSize, sectionHeaderSize uint16

func init() {
    fileHeaderSize = uint16(binary.Size(FileHeader{}))
    programHeaderSize = uint16(binary.Size(ProgramHeader{}))
    sectionHeaderSize = uint16(binary.Size(SectionHeader{}))
}

type ELFFile struct {
    header FileHeader
    programHeaders []ProgramHeader
    sectionHeaders []SectionHeader
}

func NewELFFile(class elf.Class, endianness elf.Data, abi elf.OSABI, elfType elf.Type, arch elf.Machine) *ELFFile {
    return &ELFFile{
        FileHeader{
            [4]byte{0x7f, 'E', 'L', 'F'}, // Magic
            class,
            endianness,
            elf.EV_CURRENT, // Version is always EV_CURRENT (1)
            abi,
            0, // ABI version is always zero
            [7]byte{}, // Struct padding
            elfType,
            arch,
            elf.EV_CURRENT, // File version is always EV_CURRENT (1)
            [3]byte{}, // Struct padding

            // These fields are set when the program and section tables are updated
            0, // Entry point
            uint64(fileHeaderSize), // Program header offset
            0, // Section header offset
            0, // Unused flags field
            fileHeaderSize,
            programHeaderSize,
            0, // Number of program table entries
            sectionHeaderSize,
            0, // Number of section table entries
            0,
        },
        []ProgramHeader{},
        []SectionHeader{},
    }
}

func (e *ELFFile) Bytes() []byte {
    var buf bytes.Buffer
    err := binary.Write(&buf, binary.LittleEndian, e.header)
    check(err) // TODO: handle errors properly
    return buf.Bytes()
}

type FileHeader struct {
    // ident
    magic   [4]byte
    class   elf.Class
    endianness  elf.Data
    version elf.Version
    abi elf.OSABI
    abiVersion  byte
    _ [7]byte // Pad out to 16 bytes for ident

    // Rest of struct
    elfType elf.Type
    arch elf.Machine
    fileVersion elf.Version
    _ [3]byte // Pad fileVersion out to 4 bytes

    // Change the following from uint64 to uint32 for 32 bit mode
    entryPoint uint64
    programHeaderOffset uint64
    sectionHeaderOffset uint64

    flags uint32 // Not used
    headerSize uint16
    programHeaderEntrySize uint16
    programHeaderLen uint16
    sectionHeaderEntrySize uint16
    sectionHeaderLen uint16
    shstrndx uint16
}

// 64 bit program header
type ProgramHeader struct {
    segmentType uint32
    flags uint32
    offset uint64 // File offset
    virtualAddr uint64 // Virtual memory starting address
    physicalAddr uint64 // Physical address (not relevant for most systems)
    fileSize uint64
    memSize uint64
    align uint64
}

type SectionHeader struct {
    // Placeholder
}
/*
    // Program Section Header
    phdr := ProgramHeader{
        PT_PHDR, // Type
        PF_R, // Flags
        HEADER_SIZE, // Offset from start of file (right after header)
        0x200000, // Mem offset
        0, // Physical Address (always zero)
        56, // Size (num of entries * 56)
        56,
        0, // Alignment (no idea what to set this to
    }
}*/
