package elfwriter

import (
    "encoding/binary"
    "bytes"
    "testing"
    "debug/elf"
    "reflect"
)

// Returns unequal field when structs do not match
func structEquals(a, b interface{}) (bool, string) {
    va := reflect.ValueOf(a)
    vb := reflect.ValueOf(b)

    if va.Kind() != reflect.Struct || vb.Kind() != reflect.Struct {
        panic("Cannot compare non-structs")
    }

    if va.Type() != vb.Type() {
        panic("Cannot compare structs with inconsistent fields")
    }

    for i := 0; i < va.NumField(); i++ {
        if va.Field(i).Interface() != vb.Field(i).Interface() {
            return false, va.Type().Field(i).Name
        }
    }

    return true, ""

}

func TestELFFile(t *testing.T) {
    e := NewELFFile(elf.ELFCLASS64, elf.ELFDATA2LSB, elf.ELFOSABI_FREEBSD, elf.ET_EXEC, elf.EM_X86_64)
    f, err := elf.NewFile(bytes.NewReader(e.Bytes()))

    if err != nil {
        t.Error("Can't read generated elf header:", err)
    }

    t.Run("Header", func(t *testing.T) {
        equal, mismatch := structEquals(f.FileHeader, elf.FileHeader{
            elf.ELFCLASS64,
            elf.ELFDATA2LSB,
            elf.EV_CURRENT,
            elf.ELFOSABI_FREEBSD,
            0,
            binary.LittleEndian,
            elf.ET_EXEC,
            elf.EM_X86_64,
            0, // Entry point
        })
        if !equal {
            t.Error("Header field '", mismatch, "'", "does not match")
        }
   })
}
