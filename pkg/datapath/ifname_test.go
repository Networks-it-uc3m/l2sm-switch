// ifname_test.go
package datapath

import "testing"

func TestPortAndProbeFormatting(t *testing.T) {
	ifid := Ifid{token: "abcde"}

	gotPort := ifid.Port(7)
	wantPort := "lsabcde7"
	if gotPort != wantPort {
		t.Fatalf("Port(): got %q, want %q", gotPort, wantPort)
	}

	gotProbe := ifid.Probe(7)
	wantProbe := "lsabcdeprobe7"
	if gotProbe != wantProbe {
		t.Fatalf("Probe(): got %q, want %q", gotProbe, wantProbe)
	}
}

func TestParse_Port(t *testing.T) {
	name := "lsabcde12"

	id, typ, ifid, err := Parse(name)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if id != 12 {
		t.Fatalf("Parse() id: got %d, want %d", id, 12)
	}
	if typ != TypePort {
		t.Fatalf("Parse() type: got %v, want %v", typ, TypePort)
	}
	if ifid.token != "abcde" {
		t.Fatalf("Parse() token: got %q, want %q", ifid.token, "abcde")
	}
}

func TestParse_Probe(t *testing.T) {
	name := "lsabcdeprobe42"

	id, typ, ifid, err := Parse(name)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("Parse() id: got %d, want %d", id, 42)
	}
	if typ != TypeProbe {
		t.Fatalf("Parse() type: got %v, want %v", typ, TypeProbe)
	}
	if ifid.token != "abcde" {
		t.Fatalf("Parse() token: got %q, want %q", ifid.token, "abcde")
	}
}

func TestParse_NotManaged(t *testing.T) {
	_, typ, ifid, err := Parse("eth0")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if typ != TypeUnknown {
		t.Fatalf("type: got %v, want %v", typ, TypeUnknown)
	}
	if ifid.token != "" {
		t.Fatalf("token: got %q, want empty", ifid.token)
	}
}

func TestParse_TooShort(t *testing.T) {
	// len("ls")+TOKEN_LEN+1 == 2+5+1 == 8
	_, typ, _, err := Parse("lsabcde") // length 7, too short
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if typ != TypeUnknown {
		t.Fatalf("type: got %v, want %v", typ, TypeUnknown)
	}
}

func TestParse_InvalidID(t *testing.T) {
	tests := []string{
		"lsabcdeX",
		"lsabcdeprobeX",
		"lsabcdeprobe", // empty id after "probe"
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			_, typ, _, err := Parse(name)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if typ != TypeUnknown {
				t.Fatalf("type: got %v, want %v", typ, TypeUnknown)
			}
		})
	}
}

func TestIsManaged(t *testing.T) {
	if !IsManaged("lsanything0") {
		t.Fatalf("IsManaged(): expected true for ls* name")
	}
	if IsManaged("eth0") {
		t.Fatalf("IsManaged(): expected false for non-ls name")
	}

	ifid1 := New("zzzzz")
	ifid2 := New("fsdfasd")
	inter := ifid1.Port(10)
	inter2 := ifid2.Port(10)
	if !ifid1.IsManaged(inter) {
		t.Fatalf("(*Ifid).IsManaged(): expected true for matching token")
	}
	if ifid1.IsManaged(inter2) {
		t.Fatalf("(*Ifid).IsManaged(): expected false for non-matching token")
	}
}
