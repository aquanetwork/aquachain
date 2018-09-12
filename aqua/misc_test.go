package aqua

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecodeExtra(t *testing.T) {
	b := makeExtraData(nil)
	fmt.Printf("extra: %s\n", gohex(b))
	version, _, extra, err := DecodeExtraData(makeExtraData(nil))
	if err != nil {
		fmt.Println("got err:", err)
		t.Fail()
	}
	fmt.Println("version:", version)
	fmt.Println("extra:", string(extra[6:]))
}

func TestDecodeExtra2(t *testing.T) {
	var (
		wantVersion = [3]uint8{1, 7, 7}
		b           = []byte{0xd4, 0x83, 0x1, 0x7, 0x7, 0x89, 0x61, 0x71, 0x75, 0x61, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x85, 0x6c, 0x69, 0x6e, 0x75, 0x7}
	)
	version, osname, extra, err := DecodeExtraData(b)
	if err != nil {
		t.Log("err non-nil", err)
		t.FailNow()
	}
	fmt.Println("Detected OS:", osname)
	if version[1] != wantVersion[1] {
		t.Log("version mismatch:", version, "wanted:", wantVersion)
		t.Fail()
	}
	if 0 != bytes.Compare(extra, b) {
		t.Log("extra mismatch:", gohex(extra), "wanted:", gohex(b))
		t.Fail()
	}

	fmt.Println("extra:", string(b[6:]))

}

func gohex(b []byte) (s string) {
	if len(b) == 0 {
		return "nil"
	}
	for i := range b {
		if len(b)-1 == i {
			s += fmt.Sprintf("0x%x", b[i])
			break
		}
		s += fmt.Sprintf("0x%x, ", b[i])
	}
	return s
}
