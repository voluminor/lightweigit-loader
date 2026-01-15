package tests

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"encoding/gob"
	"hash/crc32"
	"io"
	"strings"
	"testing"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

type Sample struct {
	A int
	B string
}

func safeModTypeString(m target.ModType) (s string, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	return m.String(), true
}

func findKnownModType(t *testing.T) target.ModType {
	t.Helper()

	for i := 0; i < 256; i++ {
		m := target.ModType(byte(i))
		s, ok := safeModTypeString(m)
		if !ok {
			continue
		}
		if s != "unknown" {
			return m
		}
	}

	t.Skip("no ModType found with String() != \"unknown\"")
	return 0
}

func findUnknownModType(t *testing.T) (target.ModType, bool) {
	t.Helper()

	for i := 0; i < 256; i++ {
		m := target.ModType(byte(i))
		s, ok := safeModTypeString(m)
		if !ok {
			continue
		}
		if s == "unknown" {
			return m, true
		}
	}

	return 0, false
}

func encodeGob(t *testing.T, v any) []byte {
	t.Helper()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		t.Fatalf("gob.Encode error: %v", err)
	}
	return buf.Bytes()
}

func decompressFlate(t *testing.T, compressed []byte) []byte {
	t.Helper()

	r := flate.NewReader(bytes.NewReader(compressed))
	defer r.Close()

	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		t.Fatalf("flate decompress error: %v", err)
	}
	return out.Bytes()
}

func makePacket(t *testing.T, m target.ModType, rawGob []byte) []byte {
	t.Helper()

	crc := crc32.ChecksumIEEE(rawGob)
	var crcBuf [4]byte
	binary.LittleEndian.PutUint32(crcBuf[:], crc)

	var comp bytes.Buffer
	w, err := flate.NewWriter(&comp, flate.BestCompression)
	if err != nil {
		t.Fatalf("flate.NewWriter error: %v", err)
	}
	if _, err := w.Write(rawGob); err != nil {
		t.Fatalf("flate write error: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("flate close error: %v", err)
	}

	var out bytes.Buffer
	out.WriteByte(byte(m))
	out.Write(comp.Bytes())
	out.Write(crcBuf[:])

	return out.Bytes()
}

// //

func TestMarshalUnmarshal_RoundTrip(t *testing.T) {
	m := findKnownModType(t)

	in := Sample{A: 42, B: "hello"}
	data := lightweigit.Marshal(m, &in)

	var out Sample
	gotM, err := lightweigit.Unmarshal(data, &out)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if gotM != m {
		t.Fatalf("unexpected mod type: got=%v want=%v", gotM, m)
	}

	if out != in {
		t.Fatalf("unexpected decoded value: got=%+v want=%+v", out, in)
	}
}

func TestUnmarshal_NotEnoughData(t *testing.T) {
	var out Sample
	_, err := lightweigit.Unmarshal([]byte{1, 2, 3}, &out)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not enough data") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshal_UnknownModType(t *testing.T) {
	unk, ok := findUnknownModType(t)
	if !ok {
		t.Skip("no ModType value found returning String() == \"unknown\"")
	}

	var out Sample
	data := []byte{byte(unk), 0, 0, 0, 0}

	gotM, err := lightweigit.Unmarshal(data, &out)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if gotM != unk {
		t.Fatalf("unexpected mod type: got=%v want=%v", gotM, unk)
	}
	if !strings.Contains(err.Error(), "unknown mod type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshal_InvalidChecksum(t *testing.T) {
	m := findKnownModType(t)

	in := Sample{A: 7, B: "checksum"}
	data := lightweigit.Marshal(m, &in)

	if len(data) < 6 {
		t.Fatalf("marshal returned too short data: %d bytes", len(data))
	}

	idx := len(data) / 2
	if idx <= 0 {
		idx = 1
	}
	if idx >= len(data)-4 {
		idx = 1
	}
	if idx == 0 {
		idx = 1
	}
	data[idx] ^= 0xFF

	var out Sample
	_, err := lightweigit.Unmarshal(data, &out)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid checksum") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshal_GobDecodeError_WithValidCRC(t *testing.T) {
	m := findKnownModType(t)

	raw := []byte("this is not a gob stream")
	data := makePacket(t, m, raw)

	var out Sample
	gotM, err := lightweigit.Unmarshal(data, &out)
	if gotM != m {
		t.Fatalf("unexpected mod type: got=%v want=%v", gotM, m)
	}
	if err == nil {
		t.Fatalf("expected gob decode error, got nil")
	}

	if strings.Contains(err.Error(), "invalid checksum") {
		t.Fatalf("expected decode error, got checksum error: %v", err)
	}
}

func TestMarshal_AppendsCRCOfGobBytes_AndCompressedPayloadMatches(t *testing.T) {
	m := findKnownModType(t)

	in := Sample{A: 100, B: "crc-check"}

	rawGob := encodeGob(t, in)
	wantCRC := crc32.ChecksumIEEE(rawGob)

	data := lightweigit.Marshal(m, &in)

	if data[0] != byte(m) {
		t.Fatalf("unexpected mod type byte: got=%d want=%d", data[0], byte(m))
	}

	gotCRC := binary.LittleEndian.Uint32(data[len(data)-4:])
	if gotCRC != wantCRC {
		t.Fatalf("unexpected crc: got=%d want=%d", gotCRC, wantCRC)
	}

	compressed := data[1 : len(data)-4]
	gotRaw := decompressFlate(t, compressed)

	if !bytes.Equal(gotRaw, rawGob) {
		t.Fatalf("decompressed payload != gob payload")
	}
}
