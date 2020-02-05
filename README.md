# elfwriter [WIP]
A go library for writing ELF data using the built in [`elf/debug`](https://golang.org/pkg/debug/elf/) package.

# Usage
`elfwriter` only has one public method, `WriteElf`:

```
// shrtrndx = section header table index of the entry associated with the section name string table
func WriteElf(f *elf.File, w io.WriteSeeker, programTableOffset, sectionTableOffset uint64, shstrndx uint16) error
```

It will write a complete ELF file to `w` with the given ELF data.
