// A go library for writing ELF data using the built in [`elf/debug`](https://golang.org/pkg/debug/elf/) package.
package main

import (
    "debug/elf"
    "encoding/binary"
    "io"
    "os"
)

type elfN uint64

func createBinaryWriter(w io.Writer, bo binary.ByteOrder, x32 bool) func(...interface{}) error {
    return func(data ...interface{}) error {
        for _, v := range data {
            _, isElfN := v.(elfN)
            var err error

            if x32 && isElfN {
                err = binary.Write(w, bo, uint32(v.(elfN)))
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
func WriteElf(f *elf.File, w io.WriteSeeker, programTableOffset, sectionTableOffset uint64) error {
    // Detect 32/64 bit and byteorder
    x32 := f.FileHeader.Class == elf.ELFCLASS32
    write := createBinaryWriter(w, f.FileHeader.ByteOrder, x32)

    var ehdrSize, phdrSize uint16
    if x32 {
        ehdrSize = 52
        phdrSize = 32
    } else {
        ehdrSize = 64
        phdrSize = 56
    }

    // Write file header
    fh := f.FileHeader
    err := write(
        // Identifier
        [4]byte{0x7f, 'E', 'L', 'F'}, // Magic
        fh.Class,
        fh.Data,
        fh.Version,
        fh.OSABI,
        fh.ABIVersion,
        [7]byte{}, // Pad out the identifier to 7 bytes

        // Write rest of file header
        fh.Type,
        fh.Machine,
        uint32(elf.EV_CURRENT),
        elfN(fh.Entry),
        elfN(programTableOffset),
        elfN(0), // Section Table Offset (placeholder)
        uint32(0), // Flags (unused field)
        ehdrSize,
        phdrSize,
        uint16(len(f.Progs)),
        uint16(0), // Section Header size (depends on 32/64bit)
        uint16(0), //len(f.Sections)))
        uint16(0), // Section header name table index
    )

    if err != nil {
        return err
    }

    // Write program table & program segments
    for idx, prog := range f.Progs {
        // Seek to program table entry start
        _, err = w.Seek(int64(programTableOffset) + int64(idx)*int64(phdrSize), io.SeekStart)
        if err != nil {
            return err
        }
        ph := prog.ProgHeader

        err = write(uint32(ph.Type))
        if err != nil {
            return err
        }

        // The position of the flags struct member differs between 32 and 64 bit headers
        if !x32 {
            err = write(ph.Flags)
            if err != nil {
                return err
            }
        }
        err = write(
            elfN(ph.Off),
            elfN(ph.Vaddr),
            elfN(ph.Paddr),
            elfN(ph.Filesz),
            elfN(ph.Memsz),
        )
        if err != nil {
            return err
        }
        if x32 {
            err = write(ph.Flags)
            if err != nil {
                return err
            }
        }
        err = write(ph.Align)
        if err != nil {
            return err
        }

        // Write the segment
        _, err = w.Seek(int64(ph.Off), io.SeekStart)
        if err != nil {
            return err
        }
        _, err = io.Copy(w, prog.Open())
        if err != nil {
            return err
        }
    }

    return nil
}


// Only used for testing because the library is in very early stages
func main() {
    f, e := elf.Open("elfwriter")

    // Remove the LOAD entry in the program table that loads the elf header (used for debug)
    for i, v := range f.Progs {
        if v.Off == 0 {
            f.Progs = append(f.Progs[:i], f.Progs[i+1:]...)
        }
    }
    if e != nil {
        panic(e)
    }

    var phdroffset uint64
    if f.FileHeader.Class == elf.ELFCLASS32 {
        phdroffset = 52
    } else {
        phdroffset = 64
    }

    f2, e := os.Create("out")
    e = WriteElf(f, f2, phdroffset, 0)
    if e != nil {
        panic(e)
    }
}

