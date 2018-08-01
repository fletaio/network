package formulator

import (
	"bytes"
	"io"
	"strconv"
	"testing"
	"time"
)

func TestNode_AddrMatch(t *testing.T) {
	tests := []struct {
		name string
		n    *Node
		want string
	}{
		{
			name: "addrMatch",
			n: &Node{
				Address: "testAddrmatch",
			},
			want: "testAddrmatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			_, err := tt.n.WriteTo(w)
			if err != nil {
				panic(err)
			}

			r := bytes.NewReader(w.Bytes())
			var n Node
			_, err = n.ReadFrom(r)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n.Address != tt.want {
				t.Errorf("n.Addr = %v, want %v", n.Address, tt.want)
			}
		})
	}
}
func TestNode_TimeDecode(t *testing.T) {
	nTime := time.Now()
	tests := []struct {
		name string
		n    *Node
		want string
	}{
		{
			name: "timeDecode",
			n: &Node{
				Detected: nTime,
			},
			want: strconv.Itoa(int(nTime.UnixNano())),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			_, err := tt.n.WriteTo(w)
			if err != nil {
				panic(err)
			}

			r := bytes.NewReader(w.Bytes())
			var n Node
			_, err = n.ReadFrom(r)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if strconv.Itoa(int(n.Detected.UnixNano())) != tt.want {
				t.Errorf("n.Addr = %v, want %v", n.Address, tt.want)
			}
		})
	}
}
