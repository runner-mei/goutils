package util

import "bytes"

func NewSectionParser(parser *LineParser) *SectionParser {
	return &SectionParser{parser: parser}
}

type SectionParser struct {
	parser *LineParser
}

func (p *SectionParser) Scan() bool {
	for {
		if !p.parser.Scan() {
			return false
		}
		bs := p.parser.Bytes()
		if len(bs) == 0 {
			continue
		}

		bs = bytes.TrimSpace(bs)
		if len(bs) > 0 {
			return true
		}
	}
}

func (s *SectionParser) SetError(err error) {
	s.parser.SetError(err)
}

func (p *SectionParser) Err() error {
	return p.parser.Err()
}

func (p *SectionParser) LineNumber() int {
	return p.parser.LineNumber()
}

func (p *SectionParser) Lines() Lines {
	return Lines{parser: p.parser, isFirst: true}
}

type Lines struct {
	parser  *LineParser
	isFirst bool
}

func (p *Lines) Scan() bool {
	if p.isFirst {
		p.isFirst = false
		return true
	}

	if !p.parser.Scan() {
		return false
	}
	bs := p.parser.Bytes()
	if len(bs) == 0 {
		return false
	}

	bs = bytes.TrimSpace(bs)
	if len(bs) == 0 {
		return false
	}
	return true
}

func (p *Lines) Bytes() []byte {
	return p.parser.Bytes()
}

func (p *Lines) Text() string {
	return p.parser.Text()
}

func (p *Lines) Err() error {
	return p.parser.Err()
}

func (s *Lines) SetError(err error) {
	s.parser.SetError(err)
}

func (p *Lines) LineNumber() int {
	return p.parser.LineNumber()
}
