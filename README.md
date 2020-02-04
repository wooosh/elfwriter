# elfwriter [WIP]
A go library for writing ELF data using the built in [`elf/debug`](https://golang.org/pkg/debug/elf/) package.

# Usage
`elfwriter` only has one public method, `WriteElf`:

```
func WriteElf(f *elf.File, w io.WriteSeeker, programTableOffset, sectionTableOffset uint64) error
```

It will write a complete ELF file to `w` with the given ELF data.
