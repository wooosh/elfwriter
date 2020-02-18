// A go library for writing ELF data using the built in [`elf/debug`](https://golang.org/pkg/debug/elf/) package.
package elfwriter

import (
    "debug/elf"
    "encoding/binary"
    "io"
    "errors"
)


type ELFFile struct {
    FileHeader FileHeader
    ProgramTable []ProgramSegment
    SectionTable []SectionHeader
}

type FileHeader struct {
    // ident
    Class   elf.Class
    Endianness  elf.Data
    ABI elf.OSABI
    ABIVersion  byte

    // Rest of struct
    Type elf.Type
    Arch elf.Machine

    // Change the following from uint64 to uint32 for 32 bit mode
    EntryPoint uint64
    ProgramTableOffset uint64
    SectionTableOffset uint64

    Flags uint32 // Not used
    Shstrndx uint16
}

type ProgramSegment struct {
    Type elf.ProgType
    Flags elf.ProgFlag
    Offset uint64 // File offset
    VirtualAddr uint64 // Virtual memory starting address
    PhysicalAddr uint64 // Physical address (not relevant for most systems)
    FileSize uint64
    MemSize uint64
    Align uint64
    Data []byte
}

type SectionHeader struct {}

type elfN uint64
func createBinaryWriter(w io.Writer, bo binary.ByteOrder, x32 bool) func(...interface{}) error {
    return func(data ...interface{}) error {
        for _, v := range data {
            _, isElfN := v.(elfN)
            var err error

            if isElfN {
                if x32 {
                    err = binary.Write(w, bo, uint32(v.(elfN)))
                } else {
                    err = binary.Write(w, bo, uint64(v.(elfN)))
                }
            } else {
                err = binary.Write(w, bo, v)
            }

            if err != nil {
                return err
            }
        }
        return nil
    }
}

// WriteElf writes the given ELF info to the provided writer
func (f *ELFFile) Write(w io.WriteSeeker) error {
    fh := f.FileHeader

    // Detect 32/64 bit and byteorder
    x32 := fh.Class == elf.ELFCLASS32

    var bo binary.ByteOrder
    if fh.Endianness == elf.ELFDATA2LSB {
        bo = binary.LittleEndian
    } else if fh.Endianness == elf.ELFDATA2MSB {
        bo = binary.BigEndian
    } else {
        return errors.New("Can't detect endianness")
    }

    write := createBinaryWriter(w, bo, x32)

    var ehdrSize, phdrSize, shdrSize uint16
    if x32 {
        ehdrSize = 52
        phdrSize = 32
        shdrSize = 40
    } else {
        ehdrSize = 64
        phdrSize = 56
        shdrSize = 64
    }

    // Write file header
    err := write(
        // Identifier
        [4]byte{0x7f, 'E', 'L', 'F'}, // Magic
        fh.Class,
        fh.Endianness,
        elf.EV_CURRENT,
        fh.ABI,
        fh.ABIVersion,
        [7]byte{}, // Pad out the identifier to 7 bytes

        // Write rest of file header
        fh.Type,
        fh.Arch,
        uint32(elf.EV_CURRENT),
        elfN(fh.EntryPoint),
        elfN(fh.ProgramTableOffset),
        elfN(fh.SectionTableOffset),
        uint32(0), // Flags (unused field)
        ehdrSize,
        phdrSize,
        uint16(len(f.ProgramTable)),
        shdrSize,
        uint16(len(f.SectionTable)),
        fh.Shstrndx,
    )

    if err != nil {
        return err
    }

    err = f.writeProgramTable(w, write, x32, phdrSize)
    if err != nil {
        return err
    }

    /*
    // Section Table
    for idx, section := range f.SectionTable {
        w.Seek(int64(fh.SectionTableOffset) + int64(idx)*int64(shdrSize), io.SeekStart)
        sh := section.SectionHeader

        err = write(
            uint32(idx), // Section name table index
            sh.Type,
            elfN(sh.Flags),
            elfN(sh.Addr),
            elfN(sh.Offset),
            elfN(sh.Size),
            sh.Link,
            sh.Info,
            elfN(sh.Addralign),
            elfN(sh.Entsize),
        )
        if err != nil {
            return err
        }

        _, err = w.Seek(int64(sh.Offset), io.SeekStart)
        if err != nil {
            return err
        }
        _, err = io.Copy(w, section.Open())
        if err != nil {
            return err
        }
    }*/

    return nil
}

func (f *ELFFile) writeProgramTable(w io.WriteSeeker, write func(...interface{}) error, x32 bool, phdrSize uint16) error {
    for idx, prog := range f.ProgramTable {
        // Seek to program table entry start
        _, err := w.Seek(int64(f.FileHeader.ProgramTableOffset) + int64(idx)*int64(phdrSize), io.SeekStart)
        if err != nil {
            return err
        }

        err = write(uint32(prog.Type))
        if err != nil {
            return err
        }

        // The position of the flags struct member differs between 32 and 64 bit headers
        if !x32 {
            err = write(prog.Flags)
            if err != nil {
                return err
            }
        }
        err = write(
            elfN(prog.Offset),
            elfN(prog.VirtualAddr),
            elfN(prog.PhysicalAddr),
            elfN(prog.FileSize),
            elfN(prog.MemSize),
        )
        if err != nil {
            return err
        }
        if x32 {
            err = write(prog.Flags)
            if err != nil {
                return err
            }
        }
        err = write(prog.Align)
        if err != nil {
            return err
        }

        // Write the segment
        _, err = w.Seek(int64(prog.Offset), io.SeekStart)
        if err != nil {
            return err
        }
        _, err = w.Write(prog.Data)
        if err != nil {
            return err
        }
    }

    return nil
}
