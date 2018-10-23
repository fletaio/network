package mocknetwork

import (
	"bufio"
	"fmt"
	"testing"
)

func Test_readwrite(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		rw   *readwrite
		args args
		want string
	}{
		{
			name: "test",
			rw: &readwrite{
				data: make(chan byte),
				done: make(chan int),
			},
			args: args{p: []byte("testbyte")},
			want: "testbyte",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			reader := bufio.NewReader(tt.rw)
			writer := bufio.NewWriter(tt.rw)

			writer.Write(tt.args.p)
			go writer.Flush()
			go func() {
			}()
			b := make([]byte, 40)
			n, err := reader.Read(b)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(b), " : ", tt.want)
			if string(b) != tt.want {
				t.Errorf("readwrite.Read() = %v, want %v", n, tt.want)
			}
		})
	}
}
