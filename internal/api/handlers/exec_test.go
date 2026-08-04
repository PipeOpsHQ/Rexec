package handlers

import (
	"encoding/binary"
	"testing"
)

func TestDemuxDockerStream_Plain(t *testing.T) {
	got := demuxDockerStream([]byte("hello world\n"))
	if got.Stdout != "hello world\n" || got.Combined != "hello world\n" || got.Stderr != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestDemuxDockerStream_Multiplexed(t *testing.T) {
	var raw []byte
	appendFrame := func(streamType byte, payload string) {
		hdr := make([]byte, 8)
		hdr[0] = streamType
		binary.BigEndian.PutUint32(hdr[4:], uint32(len(payload)))
		raw = append(raw, hdr...)
		raw = append(raw, payload...)
	}
	appendFrame(1, "out1")
	appendFrame(2, "err1")
	appendFrame(1, "out2")

	got := demuxDockerStream(raw)
	if got.Stdout != "out1out2" {
		t.Fatalf("stdout = %q", got.Stdout)
	}
	if got.Stderr != "err1" {
		t.Fatalf("stderr = %q", got.Stderr)
	}
	if got.Combined != "out1out2\nerr1" {
		t.Fatalf("combined = %q", got.Combined)
	}
}

func TestDemuxDockerStream_Empty(t *testing.T) {
	got := demuxDockerStream(nil)
	if got.Stdout != "" || got.Stderr != "" || got.Combined != "" {
		t.Fatalf("got %+v", got)
	}
}
