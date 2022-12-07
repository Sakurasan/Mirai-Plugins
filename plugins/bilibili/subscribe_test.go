package bilibili

import (
	"fmt"
	"sync"
	"testing"
)

func Test_bili_getBiliSub(t *testing.T) {
	type fields struct {
		sub_bili map[string]int
		sub_map  map[string]Vlist
		Mutex    sync.Mutex
	}
	type args struct {
		mid string
	}
	tests := []struct {
		name string
		// fields  fields
		args args
		// want    *Vlist
		// wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "拳头和十三要",
			args: args{mid: "389400079"},
		},
		{
			name: "徐云",
			args: args{mid: "697166795"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &bili{
				sub_bili: make(map[string]int),
				sub_map:  make(map[string]Vlist),
				Mutex:    sync.Mutex{},
			}
			got, err := a.getBiliSub(tt.args.mid)
			if err != nil {
				t.Errorf("bili.getBiliSub() error = %v", err)
				return
			}
			fmt.Println(got)
		})
	}
}
