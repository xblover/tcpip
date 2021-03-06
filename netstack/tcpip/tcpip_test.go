// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcpip

import (
	"net"
	"testing"
)

func TestSubnetContains(t *testing.T) {
	tests := []struct {
		s    Address
		m    AddressMask
		a    Address
		want bool
	}{
		{"\xa0", "\xf0", "\x90", false},
		{"\xa0", "\xf0", "\xa0", true},
		{"\xa0", "\xf0", "\xa5", true},
		{"\xa0", "\xf0", "\xaf", true},
		{"\xa0", "\xf0", "\xb0", false},
		{"\xa0", "\xf0", "", false},
		{"\xa0", "\xf0", "\xa0\x00", false},
		{"\xc2\x80", "\xff\xf0", "\xc2\x80", true},
		{"\xc2\x80", "\xff\xf0", "\xc2\x00", false},
		{"\xc2\x00", "\xff\xf0", "\xc2\x00", true},
		{"\xc2\x00", "\xff\xf0", "\xc2\x80", false},
	}
	for _, tt := range tests {
		s, err := NewSubnet(tt.s, tt.m)
		if err != nil {
			t.Errorf("NewSubnet(%v, %v) = %v", tt.s, tt.m, err)
			continue
		}
		if got := s.Contains(tt.a); got != tt.want {
			t.Errorf("Subnet(%v).Contains(%v) = %v, want %v", s, tt.a, got, tt.want)
		}
	}
}

func TestSubnetBits(t *testing.T) {
	tests := []struct {
		a     AddressMask
		want1 int
		want0 int
	}{
		{"\x00", 0, 8},
		{"\x00\x00", 0, 16},
		{"\x36", 4, 4},
		{"\x5c", 4, 4},
		{"\x5c\x5c", 8, 8},
		{"\x5c\x36", 8, 8},
		{"\x36\x5c", 8, 8},
		{"\x36\x36", 8, 8},
		{"\xff", 8, 0},
		{"\xff\xff", 16, 0},
	}
	for _, tt := range tests {
		s := &Subnet{mask: tt.a}
		got1, got0 := s.Bits()
		if got1 != tt.want1 || got0 != tt.want0 {
			t.Errorf("Subnet{mask: %x}.Bits() = %d, %d, want %d, %d", tt.a, got1, got0, tt.want1, tt.want0)
		}
	}
}

func TestSubnetPrefix(t *testing.T) {
	tests := []struct {
		a    AddressMask
		want int
	}{
		{"\x00", 0},
		{"\x00\x00", 0},
		{"\x36", 0},
		{"\x86", 1},
		{"\xc5", 2},
		{"\xff\x00", 8},
		{"\xff\x36", 8},
		{"\xff\x8c", 9},
		{"\xff\xc8", 10},
		{"\xff", 8},
		{"\xff\xff", 16},
	}
	for _, tt := range tests {
		s := &Subnet{mask: tt.a}
		got := s.Prefix()
		if got != tt.want {
			t.Errorf("Subnet{mask: %x}.Bits() = %d want %d", tt.a, got, tt.want)
		}
	}
}

func TestSubnetCreation(t *testing.T) {
	tests := []struct {
		a    Address
		m    AddressMask
		want error
	}{
		{"\xa0", "\xf0", nil},
		{"\xa0\xa0", "\xf0", errSubnetLengthMismatch},
		{"\xaa", "\xf0", errSubnetAddressMasked},
		{"", "", nil},
	}
	for _, tt := range tests {
		if _, err := NewSubnet(tt.a, tt.m); err != tt.want {
			t.Errorf("NewSubnet(%v, %v) = %v, want %v", tt.a, tt.m, err, tt.want)
		}
	}
}

func TestRouteMatch(t *testing.T) {
	tests := []struct {
		d    Address
		m    AddressMask
		a    Address
		want bool
	}{
		{"\xc2\x80", "\xff\xf0", "\xc2\x80", true},
		{"\xc2\x80", "\xff\xf0", "\xc2\x00", false},
		{"\xc2\x00", "\xff\xf0", "\xc2\x00", true},
		{"\xc2\x00", "\xff\xf0", "\xc2\x80", false},
	}
	for _, tt := range tests {
		r := Route{Destination: tt.d, Mask: tt.m}
		if got := r.Match(tt.a); got != tt.want {
			t.Errorf("Route(%v).Match(%v) = %v, want %v", r, tt.a, got, tt.want)
		}
	}
}

func TestAddressString(t *testing.T) {
	for _, want := range []string{
		// Taken from stdlib.
		"2001:db8::123:12:1",
		"2001:db8::1",
		"2001:db8:0:1:0:1:0:1",
		"2001:db8:1:0:1:0:1:0",
		"2001::1:0:0:1",
		"2001:db8:0:0:1::",
		"2001:db8::1:0:0:1",
		"2001:db8::a:b:c:d",

		// Leading zeros.
		"::1",
		// Trailing zeros.
		"8::",
		// No zeros.
		"1:1:1:1:1:1:1:1",
		// Longer sequence is after other zeros, but not at the end.
		"1:0:0:1::1",
		// Longer sequence is at the beginning, shorter sequence is at
		// the end.
		"::1:1:1:0:0",
		// Longer sequence is not at the beginning, shorter sequence is
		// at the end.
		"1::1:1:0:0",
		// Longer sequence is at the beginning, shorter sequence is not
		// at the end.
		"::1:1:0:0:1",
		// Neither sequence is at an end, longer is after shorter.
		"1:0:0:1::1",
		// Shorter sequence is at the beginning, longer sequence is not
		// at the end.
		"0:0:1:1::1",
		// Shorter sequence is at the beginning, longer sequence is at
		// the end.
		"0:0:1:1:1::",
		// Short sequences at both ends, longer one in the middle.
		"0:1:1::1:1:0",
		// Short sequences at both ends, longer one in the middle.
		"0:1::1:0:0",
		// Short sequences at both ends, longer one in the middle.
		"0:0:1::1:0",
		// Longer sequence surrounded by shorter sequences, but none at
		// the end.
		"1:0:1::1:0:1",
	} {
		addr := Address(net.ParseIP(want))
		if got := addr.String(); got != want {
			t.Errorf("Address(%x).String() = '%s', want = '%s'", addr, got, want)
		}
	}
}
