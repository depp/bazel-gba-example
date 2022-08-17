package romimage

import (
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const (
	// maxSize is the maximum size of a ROM image, in bytes. It's technically
	// possible for cartridges to be larger but this is only found on video
	// cartridges (bank switching?)
	maxSize = 32 * 1024 * 1024

	// addrStart is the address of the beginning of the ROM image.
	addrStart = 0x8000000

	// headerLength is the length of the ROM header.
	headerLength = 0xc0

	// TitleLength is the length of the title field.
	TitleLength = 12
)

var logo = []byte{
	0x24, 0xFF, 0xAE, 0x51, 0x69, 0x9A, 0xA2, 0x21, 0x3D, 0x84, 0x82, 0x0A, 0x84, 0xE4, 0x09, 0xAD,
	0x11, 0x24, 0x8B, 0x98, 0xC0, 0x81, 0x7F, 0x21, 0xA3, 0x52, 0xBE, 0x19, 0x93, 0x09, 0xCE, 0x20,
	0x10, 0x46, 0x4A, 0x4A, 0xF8, 0x27, 0x31, 0xEC, 0x58, 0xC7, 0xE8, 0x33, 0x82, 0xE3, 0xCE, 0xBF,
	0x85, 0xF4, 0xDF, 0x94, 0xCE, 0x4B, 0x09, 0xC1, 0x94, 0x56, 0x8A, 0xC0, 0x13, 0x72, 0xA7, 0xFC,
	0x9F, 0x84, 0x4D, 0x73, 0xA3, 0xCA, 0x9A, 0x61, 0x58, 0x97, 0xA3, 0x27, 0xFC, 0x03, 0x98, 0x76,
	0x23, 0x1D, 0xC7, 0x61, 0x03, 0x04, 0xAE, 0x56, 0xBF, 0x38, 0x84, 0x00, 0x40, 0xA7, 0x0E, 0xFD,
	0xFF, 0x52, 0xFE, 0x03, 0x6F, 0x95, 0x30, 0xF1, 0x97, 0xFB, 0xC0, 0x85, 0x60, 0xD6, 0x80, 0x25,
	0xA9, 0x63, 0xBE, 0x03, 0x01, 0x4E, 0x38, 0xE2, 0xF9, 0xA2, 0x34, 0xFF, 0xBB, 0x3E, 0x03, 0x44,
	0x78, 0x00, 0x90, 0xCB, 0x88, 0x11, 0x3A, 0x94, 0x65, 0xC0, 0x7C, 0x63, 0x87, 0xF0, 0x3C, 0xAF,
	0xD6, 0x25, 0xE4, 0x8B, 0x38, 0x0A, 0xAC, 0x72, 0x21, 0xD4, 0xF8, 0x07,
}

type badaddr struct {
	name   string
	value  uint32
	expect uint32
}

func (e *badaddr) Error() string {
	return fmt.Sprintf("symbol %s has address $%x, should be $%x", e.name, e.value, e.expect)
}

func getLMA(f *elf.File, off uint64) (uint64, error) {
	for _, p := range f.Progs {
		if p.Off <= off && off < p.Off+p.Filesz {
			return off - p.Off + p.Paddr, nil
		}
	}
	return 0, fmt.Errorf("could not map offset to LMA: $%08x", off)
}

type section struct {
	lma     uint32
	size    uint32
	section *elf.Section
}

func mapSections(f *elf.File) ([]section, error) {
	var ss []section
	for _, s := range f.Sections {
		if s.Flags&elf.SHF_ALLOC == 0 || s.Type == elf.SHT_NOBITS || s.Size == 0 {
			continue
		}
		lma, err := getLMA(f, s.Offset)
		if err != nil {
			return nil, fmt.Errorf("section %q: %w", s.Name, err)
		}
		ss = append(ss, section{
			lma:     uint32(lma),
			size:    uint32(s.Size),
			section: s,
		})
	}
	return ss, nil
}

// getSize returns the size of the ROM image.
func getSize(ss []section) (uint32, error) {
	rstart := ^uint32(0)
	rend := uint32(0)
	for _, s := range ss {
		start, end := s.lma, s.lma+s.size
		if start < rstart {
			rstart = start
		}
		if end > rend {
			rend = end
		}
	}
	if rstart >= rend {
		return 0, errors.New("no program data")
	}
	if rstart < addrStart {
		return 0, fmt.Errorf("program contains data at address $%x, but the minimum address is $%x", rstart, addrStart)
	}
	n := rend - rstart
	if n > maxSize {
		return 0, fmt.Errorf("program data is too large: size=$%x, maximum=$%x", n, maxSize)
	}
	return n, nil
}

func checkSymbols(f *elf.File) error {
	// Get symbols, do sanity checks.
	syms, err := f.Symbols()
	if err != nil {
		return err
	}
	const (
		symStart         = "_start"
		symRomHeaderEnd  = "rom_header_end"
		addrRomHeaderEnd = addrStart + headerLength
	)
	var start, romHeaderEnd uint32
	symnames := map[string]*uint32{
		symStart:        &start,
		symRomHeaderEnd: &romHeaderEnd,
	}
	for _, sym := range syms {
		if v := symnames[sym.Name]; v != nil {
			*v = uint32(sym.Value)
			delete(symnames, sym.Name)
		}
	}
	if len(symnames) != 0 {
		var n []string
		for k := range symnames {
			n = append(n, k)
		}
		return fmt.Errorf("required symbols do not exist: %s", strings.Join(n, ", "))
	}
	if start != addrStart {
		return &badaddr{symStart, start, addrStart}
	}
	if romHeaderEnd != addrRomHeaderEnd {
		return &badaddr{symRomHeaderEnd, romHeaderEnd, addrRomHeaderEnd}
	}
	return nil
}

func readProgram(progname string) ([]byte, error) {
	f, err := elf.Open(progname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Sanity checks.
	if f.Class != elf.ELFCLASS32 {
		return nil, fmt.Errorf("program ELF class is %v, expected %v", f.Class, elf.ELFCLASS32)
	}
	if f.ByteOrder != binary.LittleEndian {
		return nil, errors.New("program byte order is not little-endian")
	}
	if f.Type != elf.ET_EXEC {
		return nil, fmt.Errorf("program ELF type is %v, expected %v", f.Type, elf.ET_EXEC)
	}
	if f.Machine != elf.EM_ARM {
		return nil, fmt.Errorf("program machine is %v, expect %v", f.Machine, elf.EM_ARM)
	}

	// Create ROM image.
	ss, err := mapSections(f)
	if err != nil {
		return nil, err
	}
	size, err := getSize(ss)
	if err != nil {
		return nil, err
	}
	if err := checkSymbols(f); err != nil {
		return nil, err
	}
	data := make([]byte, size)
	for _, s := range ss {
		start := s.lma - addrStart
		end := start + s.size
		if _, err := s.section.ReadAt(data[start:end], 0); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func checksum(d []byte) byte {
	val := ^uint32(0x18)
	for _, b := range d {
		val -= uint32(b)
	}
	return byte(val)
}

// Info contains metadata for a GBA ROM image.
type Info struct {
	Title string
}

// Make creates a GBA ROM image from an ELF binary.
func Make(progname string, info *Info) ([]byte, error) {
	p, err := readProgram(progname)
	if err != nil {
		return nil, err
	}
	if len(p) < headerLength {
		old := p
		p = make([]byte, headerLength)
		copy(p, old)
	}
	_ = p[4:headerLength]
	for i := 4; i < headerLength; i++ {
		p[i] = 0
	}
	copy(p[4:], logo)
	copy(p[0xa0:0xa0+TitleLength], info.Title)
	// 0xac:0xb0: game code
	// 0xb0:0xb2: maker code (01 is Nintendo)
	p[0xb0] = '0'
	p[0xb1] = '1'
	p[0xb2] = 0x96
	p[0xbd] = checksum(p[0xa0:0xbd])
	return p, nil
}
