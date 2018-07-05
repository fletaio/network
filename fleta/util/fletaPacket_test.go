package util

import (
	"reflect"
	"testing"
)

func TestFletaPacket_validate(t *testing.T) {
	tests := []struct {
		name    string
		fp      *FletaPacket
		fp2     *FletaPacket
		wantErr bool
	}{
		{
			name: "", fp: &FletaPacket{
				Command:     "testtest",
				Compression: false,
				Content:     "content",
			}, wantErr: false,
		},
		{
			name: "", fp: &FletaPacket{
				Command:     "testtes",
				Compression: false,
				Content:     "Command length is 8",
			}, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fp.validate(); (err != nil) != tt.wantErr {
				t.Errorf("FletaPacket.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFletaPacket_Packet(t *testing.T) {
	tests := []struct {
		name    string
		fp      *FletaPacket
		fp2     *FletaPacket
		wantErr bool
	}{
		{
			name: "", fp: &FletaPacket{
				Command:     "testtest",
				Compression: false,
				Content:     "Compression test Compression test Compression test Compression test Compression test Compression test Compression test Compression test Compression test ",
			}, fp2: &FletaPacket{
				Command:     "testtest",
				Compression: true,
				Content:     "Compression test Compression test Compression test Compression test Compression test Compression test Compression test Compression test Compression test ",
			}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fp.Packet()
			if (err != nil) != tt.wantErr {
				t.Errorf("FletaPacket.Packet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got2, err := tt.fp2.Packet()
			if (err != nil) != tt.wantErr {
				t.Errorf("FletaPacket.Packet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) < len(got2) {
				t.Errorf("FletaPacket Compression shoud be less size, fp : %v, fp2 %v", len(got), len(got2))
			}
		})
	}
}

func TestFletaPacketToStruct(t *testing.T) {
	type args struct {
		buf []byte
	}
	tests := []struct {
		name string
		fp   *FletaPacket
		want string
	}{
		{
			name: "", fp: &FletaPacket{
				Command:     "testtest",
				Compression: false,
				Content:     "[PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test",
			},
			want: "[PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test",
		},
		{
			name: "", fp: &FletaPacket{
				Command:     "testtest",
				Compression: true,
				Content:     "[PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test",
			},
			want: "[PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct testPacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test PacketToStruct test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := tt.fp.Packet()
			if err != nil {
				t.Errorf("FletaPacket.Packet() error = %v", err)
				return
			}

			t.Logf("%d\n", len(p))

			if got := FletaPacketToStruct(p).Content; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FletaPacketToStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
