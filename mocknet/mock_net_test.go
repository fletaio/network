package mocknet

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
)

func TestListenAndDial(t *testing.T) {
	type args struct {
		networkType string
		address     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "getConn",
			args:    args{networkType: "tcp", address: ""},
			want:    "testsend",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			var address string

			t.Log("Listen go")
			conn, err := Listen("tcp", ":3000")
			if err != nil {
				panic(err)
			}
			ls := conn
			t.Log("ls :", ls)
			address = ls.Addr().String()

			endChan := make(chan bool)
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				wg.Done()
				t.Log("address :", address)
				t.Log("befor Accept")
				conn2, err := ls.Accept()
				t.Log("after Accept")
				if err != nil {
					panic(err)
				}

				buf := make([]byte, 0, 4096) // big buffer

				for {
					tmp := make([]byte, 256) // using small tmo buffer for demonstrating
					n, err := conn2.Read(tmp)
					if err != nil {
						if err != io.EOF {
							fmt.Println("read error:", err)
						}
						break
					}

					buf = append(buf, tmp[:n]...)

					t.Log("total size:", len(buf))
					t.Log(string(buf))
				}

				got = string(buf)
				endChan <- true
			}()

			go func() {
				wg.Done()
				t.Log("Dial go address : ", address)
				conn, err := Dial(tt.args.networkType, address, "localhost:3000")
				if (err != nil) != tt.wantErr {
					t.Errorf("Dial() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				defer conn.Close()
				conn.Write([]byte(tt.want))
			}()
			wg.Wait()

			<-endChan

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dial() = %v, want %v", got, tt.want)
			}
		})
	}
}
