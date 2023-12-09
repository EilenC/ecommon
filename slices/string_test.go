package slices

import (
	"reflect"
	"testing"
)

func TestRemoveStringDuplicateUseMap(t *testing.T) {
	type args struct {
		list []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "number",
			args: args{
				list: []string{"1", "2", "2", "3"},
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "abc",
			args: args{
				list: []string{"A", "B", "B", "C"},
			},
			want: []string{"A", "B", "C"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveStringDuplicateUseMap(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveStringDuplicateUseMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveStringDuplicateUseCopy(t *testing.T) {
	type args struct {
		list []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "number",
			args: args{
				list: []string{"1", "2", "2", "3"},
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "abc",
			args: args{
				list: []string{"A", "B", "B", "C"},
			},
			want: []string{"A", "B", "C"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveStringDuplicateUseCopy(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveStringDuplicateUseCopy() = %v, want %v", got, tt.want)
			}
		})
	}
}
