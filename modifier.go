// Modifier instance is provided to milter handlers to modify email messages

package milter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/textproto"
)

// postfix wants LF lines endings. Using CRLF results in double CR sequences.
func crlfToLF(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte{'\r', '\n'}, []byte{'\n'})
}

// Modifier provides access to Macros, Headers and Body data to callback handlers. It also defines a
// number of functions that can be used by callback handlers to modify processing of the email message
type Modifier interface {
	// AddRecipient appends a new envelope recipient for current message
	AddRecipient(r string) error

	// DeleteRecipient removes an envelope recipient address from message
	DeleteRecipient(r string) error

	// ReplaceBody substitutes message body with provided body
	ReplaceBody(body []byte) error

	// AddHeader appends a new email message header the message
	AddHeader(name, value string) error

	// ChangeHeader replaces the header at the specified position with a new one.
	// The index is per name.
	ChangeHeader(index int, name, value string) error

	// InsertHeader inserts the header at the specified position
	InsertHeader(index int, name, value string) error

	// Quarantine a message by giving a reason to hold it
	Quarantine(reason string) error

	// ChangeFrom replaces the FROM envelope header with a new one
	ChangeFrom(value string) error

	// GetMacros returns Macros
	GetMacros() map[string]string

	// GetHeaders returns Headers
	GetHeaders() textproto.MIMEHeader
}

type modifier struct {
	Macros  map[string]string
	Headers textproto.MIMEHeader

	writePacket func(*Message) error
}

// AddRecipient appends a new envelope recipient for current message
func (m *modifier) AddRecipient(r string) error {
	data := []byte(fmt.Sprintf("<%s>", r) + null)
	return m.writePacket(NewResponse('+', data).Response())
}

// DeleteRecipient removes an envelope recipient address from message
func (m *modifier) DeleteRecipient(r string) error {
	data := []byte(fmt.Sprintf("<%s>", r) + null)
	return m.writePacket(NewResponse('-', data).Response())
}

// ReplaceBody substitutes message body with provided body
func (m *modifier) ReplaceBody(body []byte) error {
	body = crlfToLF(body)
	return m.writePacket(NewResponse('b', body).Response())
}

// AddHeader appends a new email message header the message
func (m *modifier) AddHeader(name, value string) error {
	var buffer bytes.Buffer
	buffer.WriteString(name + null)
	buffer.Write(crlfToLF([]byte(value)))
	buffer.WriteString(null)
	return m.writePacket(NewResponse('h', buffer.Bytes()).Response())
}

// Quarantine a message by giving a reason to hold it
func (m *modifier) Quarantine(reason string) error {
	return m.writePacket(NewResponse('q', []byte(reason+null)).Response())
}

// ChangeHeader replaces the header at the specified position with a new one.
// The index is per name.
func (m *modifier) ChangeHeader(index int, name, value string) error {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, uint32(index)); err != nil {
		return err
	}
	buffer.WriteString(name + null)
	buffer.Write(crlfToLF([]byte(value)))
	buffer.WriteString(null)
	return m.writePacket(NewResponse('m', buffer.Bytes()).Response())
}

// InsertHeader inserts the header at the specified position
func (m *modifier) InsertHeader(index int, name, value string) error {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, uint32(index)); err != nil {
		return err
	}
	buffer.WriteString(name + null)
	buffer.Write(crlfToLF([]byte(value)))
	buffer.WriteString(null)
	return m.writePacket(NewResponse('i', buffer.Bytes()).Response())
}

// ChangeFrom replaces the FROM envelope header with a new one
func (m *modifier) ChangeFrom(value string) error {
	data := []byte(value + null)
	return m.writePacket(NewResponse('e', data).Response())
}

// GetMacros returns Macros
func (m *modifier) GetMacros() map[string]string {
	return m.Macros
}

// GetHeaders returns Headers
func (m *modifier) GetHeaders() textproto.MIMEHeader {
	return m.Headers
}

// newModifier creates a new Modifier instance from milterSession
func newModifier(s *milterSession) Modifier {
	return &modifier{
		Macros:      s.macros,
		Headers:     s.headers,
		writePacket: s.WritePacket,
	}
}
