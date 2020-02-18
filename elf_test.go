package elfwriter

import (
    "io"
    "encoding/binary"
    "bytes"
    "testing"
    "debug/elf"
    "reflect"
    "errors"
)

// https://stackoverflow.com/questions/45836767/using-an-io-writeseeker-without-a-file-in-go
// Writer-Seeker buffer
type wsbuffer struct {
    buf []byte
    pos int
}

func (m *wsbuffer) Write(p []byte) (n int, err error) {
    minCap := m.pos + len(p)
    if minCap > cap(m.buf) { // Make sure buf has enough capacity:
        buf2 := make([]byte, len(m.buf), minCap+len(p)) // add some extra
        copy(buf2, m.buf)
        m.buf = buf2
    }
    if minCap > len(m.buf) {
        m.buf = m.buf[:minCap]
    }
    copy(m.buf[m.pos:], p)
    m.pos += len(p)
    return len(p), nil
}

func (m *wsbuffer) Seek(offset int64, whence int) (int64, error) {
    newPos, offs := 0, int(offset)
    switch whence {
    case io.SeekStart:
        newPos = offs
    case io.SeekCurrent:
        newPos = m.pos + offs
    case io.SeekEnd:
        newPos = len(m.buf) + offs
    }
    if newPos < 0 {
        return 0, errors.New("negative result pos")
    }
    m.pos = newPos
    return int64(newPos), nil
}

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
    fh := FileHeader{
        elf.ELFCLASS64,
        elf.ELFDATA2LSB,
        elf.ELFOSABI_FREEBSD,
        0,
        elf.ET_EXEC,
        elf.EM_X86_64,
        0,
        0,
        0,
        0,
        0,
    }

    e := ELFFile{
        fh,
        []ProgramSegment{},
        []SectionHeader{},
    }


    var buf wsbuffer
    e.Write(&buf)

    f, err := elf.NewFile(bytes.NewReader(buf.buf))

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
