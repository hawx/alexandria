package mobi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type Metadata struct {
	Title       string
	Creator     string
	Publisher   string
	Description string
	Isbn        string
	Subject     []string
	Date        string
	Contributor string
	Rights      string
	SubjectCode string
	Type        string
	Source      string
}

type Mobi interface {
	Metadata() (Metadata, error)
	Close() error
}

type mobi struct {
	r readSeekCloser
	c io.Closer
}

type readSeekCloser interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

func Open(file readSeekCloser) (Mobi, error) {
	return &mobi{file, file}, nil
}

func (m *mobi) Metadata() (Metadata, error) {
	const size = 4096
	metadata := Metadata{}
	buf := make([]byte, size)

	m.r.ReadAt(buf, 0)
	p := bytes.NewBuffer(buf)

	var header struct {
		_               [34]byte //  0
		Version         uint16   // 34
		CreationDate    uint32   // 36
		_               [12]byte // 40
		AppInfoID       uint32   // 52
		SortInfoID      uint32   // 56
		Type            [4]byte  // 60
		Creator         [4]byte  // 64
		_               [8]byte  // 68
		NumberOfRecords uint16   // 76
	}

	binary.Read(p, binary.BigEndian, &header)
	if string(header.Creator[:]) != "MOBI" || string(header.Type[:]) != "BOOK" {
		return Metadata{}, errors.New("not a book")
	}

	// 78: 8 * number of records
	p.Next(8 * int(header.NumberOfRecords))

	// xx: skip 2 zero bytes
	p.Next(2)

	var mobiHeader struct {
		_              [16]byte //   0  -- PalmDOC header
		Identifier     [4]byte  //  16
		Length         uint32   //  20  -- includes previous 4 bytes
		_              [60]byte //  24
		FullNameOffset uint32   //  84
		FullNameLength uint32   //  88
		_              [36]byte //  92
		ExthFlags      [4]byte  // 128
	}

	binary.Read(p, binary.BigEndian, &mobiHeader)
	if string(mobiHeader.Identifier[:]) != "MOBI" {
		return Metadata{}, errors.New("not a book")
	}

	p.Next(int(mobiHeader.Length) - 128 + 12) // +12 to skip PalmDOC header as well

	hasExthHeader := mobiHeader.ExthFlags[3]&0x40 != 0
	if hasExthHeader {
		var exthHeader struct {
			Identifier [4]byte // 0
			Length     uint32  // 4  -- includes previous 4 bytes
			Count      uint32  // 8
		}

		binary.Read(p, binary.BigEndian, &exthHeader)

		if string(exthHeader.Identifier[:]) != "EXTH" {
			return Metadata{}, errors.New("expected EXTH header")
		}

		type exthRecord struct {
			Type   uint32 // 0
			Length uint32 // 4
			Data   []byte // 8:Length-8
		}

		exthRecords := make([]exthRecord, exthHeader.Count)

		for i := 0; i < len(exthRecords); i++ {
			binary.Read(p, binary.BigEndian, &exthRecords[i].Type)
			binary.Read(p, binary.BigEndian, &exthRecords[i].Length)
			recordLength := exthRecords[i].Length
			if exthRecords[i].Length > 8 {
				recordLength -= 8
			}
			exthRecords[i].Data = make([]byte, recordLength)
			binary.Read(p, binary.BigEndian, &exthRecords[i].Data)
		}

		for _, rec := range exthRecords {
			val := string(rec.Data[:])
			switch rec.Type {
			case 100:
				metadata.Creator = val
			case 101:
				metadata.Publisher = val
			case 103:
				metadata.Description = val
			case 104:
				metadata.Isbn = val
			case 105:
				metadata.Subject = append(metadata.Subject, val)
			case 106:
				metadata.Date = val
			case 108:
				metadata.Contributor = val
			case 109:
				metadata.Rights = val
			case 110:
				metadata.SubjectCode = val
			case 111:
				metadata.Type = val
			case 112:
				metadata.Source = val
			}
		}

		// Skip padding
		for i := 0; i < 4; i++ {
			b, _ := p.ReadByte()
			if b != 0x00 {
				p.UnreadByte()
				break
			}
		}
	}

	name := make([]byte, mobiHeader.FullNameLength)
	binary.Read(p, binary.BigEndian, &name)
	metadata.Title = string(name[:])

	return metadata, nil
}

func (m *mobi) Close() error {
	return m.c.Close()
}
