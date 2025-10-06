package common

import (
	"encoding/json"
	"fmt"
)

type LineCol struct {
	LineNo int // The starting line number of the token
	ColNo  int // The starting column number of the token
}

type Span struct {
	StartLine   int // The starting line number of the token
	StartColumn int // The starting column number of the token
	EndLine     int // The ending line number of the token
	EndColumn   int // The ending column number of the token
}

func (x *Span) SpanString() string {
	return fmt.Sprintf("%d %d %d %d", x.StartLine, x.StartColumn, x.EndLine, x.EndColumn)
}

func (x *LineCol) SpanString(lineCol LineCol) string {
	span := Span{
		StartLine:   x.LineNo,
		StartColumn: x.ColNo,
		EndLine:     lineCol.LineNo,
		EndColumn:   lineCol.ColNo,
	}
	return span.SpanString()
}

func (x *LineCol) Span(lineCol LineCol) Span {
	return Span{
		StartLine:   x.LineNo,
		StartColumn: x.ColNo,
		EndLine:     lineCol.LineNo,
		EndColumn:   lineCol.ColNo,
	}
}

func (x *Span) ToSpan(y *Span) *Span {
	return &Span{
		StartLine:   x.StartLine,
		StartColumn: x.StartColumn,
		EndLine:     y.EndLine,
		EndColumn:   y.EndColumn,
	}
}

func (x *Span) MergeSpan(y *Span) Span {
	if y == nil {
		return Span{}
	}
	sofar := *x
	if sofar.StartLine > y.StartLine || (sofar.StartLine == y.StartLine && sofar.StartColumn > y.StartColumn) {
		sofar.StartLine = y.StartLine
		sofar.StartColumn = y.StartColumn
	}
	if sofar.EndLine < y.EndLine || (sofar.EndLine == y.EndLine && sofar.EndColumn < y.EndColumn) {
		sofar.EndLine = y.EndLine
		sofar.EndColumn = y.EndColumn
	}
	return sofar
}

// MarshalJSON implements custom JSON marshaling for Span.
func (s Span) MarshalJSON() ([]byte, error) {
	arr := [4]int{s.StartLine, s.StartColumn, s.EndLine, s.EndColumn}
	return json.Marshal(arr)
}

// UnmarshalJSON implements custom JSON unmarshaling for Span.
func (s *Span) UnmarshalJSON(data []byte) error {
	var arr [4]int
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	s.StartLine = arr[0]
	s.StartColumn = arr[1]
	s.EndLine = arr[2]
	s.EndColumn = arr[3]
	return nil
}
