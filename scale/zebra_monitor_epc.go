package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	testEPCMu     sync.Mutex
	testEPCLastNS int64
	testEPCSeq    uint32
	testEPCSalt   uint32 = newEPCSalt()
)

func safeText(fallback, v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func sanitizeZPLText(v string) string {
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.ReplaceAll(v, "^", " ")
	v = strings.ReplaceAll(v, "~", " ")
	return strings.TrimSpace(v)
}

func buildRFIDEncodeCommand(epc, qtyText, itemName string) (string, error) {
	norm, err := normalizeEPC(epc)
	if err != nil {
		return "", err
	}

	qty := sanitizeZPLText(strings.TrimSpace(qtyText))
	if qty == "" {
		qty = "- kg"
	}
	item := sanitizeZPLText(strings.TrimSpace(itemName))
	if item == "" {
		item = "-"
	}

	return "^XA\n" +
		"^LH0,0\n" +
		"^MMT\n" +
		"^RS8,,,1,N\n" +
		fmt.Sprintf("^RFW,H,,,A^FD%s^FS\n", norm) +
		"^FO8,30^A0N,24,20^FB760,1,0,L,0\n" +
		fmt.Sprintf("^FDITEM: %s^FS\n", item) +
		"^FO8,74^A0N,32,28\n" +
		fmt.Sprintf("^FDQTY: %s^FS\n", qty) +
		"^FO8,126^A0N,24,20^FB760,1,0,L,0\n" +
		fmt.Sprintf("^FDEPC: %s^FS\n", sanitizeZPLText(norm)) +
		"^FO8,162^BY2,2,44^BCN,44,N,N,N\n" +
		fmt.Sprintf("^FD%s^FS\n", sanitizeZPLText(norm)) +
		"^PQ1\n" +
		"^XZ\n", nil
}

func normalizeEPC(epc string) (string, error) {
	v := strings.ToUpper(strings.TrimSpace(epc))
	v = strings.TrimPrefix(v, "0X")
	v = strings.ReplaceAll(v, " ", "")
	v = strings.ReplaceAll(v, "-", "")

	if v == "" {
		return "", errors.New("epc bo'sh")
	}
	if !zebraHexOnlyRegex.MatchString(v) {
		return "", errors.New("epc faqat hex bo'lishi kerak")
	}
	if len(v)%2 != 0 {
		return "", errors.New("epc uzunligi juft bo'lishi kerak")
	}
	if len(v) < 8 || len(v) > 64 {
		return "", errors.New("epc uzunligi 8..64 oralig'ida bo'lsin")
	}
	return v, nil
}

func generateTestEPC(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	ns := t.UnixNano()

	testEPCMu.Lock()
	if ns != testEPCLastNS {
		testEPCLastNS = ns
		testEPCSeq = 0
	} else {
		testEPCSeq++
	}
	seq := testEPCSeq
	salt := testEPCSalt
	testEPCMu.Unlock()

	return formatEPC24(ns, seq, salt)
}

func formatEPC24(ns int64, seq, salt uint32) string {
	atom := uint32((uint64(ns) / 1_000) & 0xFFFFFFFF)
	tail := atom ^ bits.RotateLeft32(uint32(ns), 13) ^ bits.RotateLeft32(seq, 7) ^ salt
	tail |= 1
	return fmt.Sprintf("30%014X%08X", uint64(ns)&0x00FFFFFFFFFFFFFF, tail)
}

func newEPCSalt() uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err == nil {
		return binary.BigEndian.Uint32(b[:]) | 1
	}
	fallback := uint32(time.Now().UnixNano()) ^ (uint32(os.Getpid()) << 16)
	return fallback | 1
}

func inferVerify(line1, line2, expected string) string {
	line1 = strings.TrimSpace(strings.Trim(line1, "\""))
	line2 = strings.TrimSpace(strings.Trim(line2, "\""))
	text := strings.ToUpper(strings.TrimSpace(line1 + " " + line2))

	if text == "" || text == "-" {
		return "UNKNOWN"
	}
	if strings.Contains(strings.ToLower(text), "no tag") {
		return "NO TAG"
	}
	if strings.Contains(strings.ToLower(text), "error") {
		return "ERROR"
	}

	expected = strings.ToUpper(strings.TrimSpace(expected))
	if expected != "" {
		if strings.Contains(strings.ReplaceAll(text, " ", ""), expected) {
			return "MATCH"
		}
		return "MISMATCH"
	}

	return "OK"
}
