package ecommon

import "testing"

func TestGetAttachmentName(t *testing.T) {
	type args struct {
		path string
		sep  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Contains custom attachment names",
			args: args{
				path: "https://baike.seekhill.com/uploads/202106/1624354355xY7cLkuE_s.png|golang.png",
			},
			want: "golang.png",
		},
		{
			name: "Contains custom attachment names",
			args: args{
				path: "https://baike.seekhill.com/uploads/202106/1624354355xY7cLkuE_s.png-golang.png",
				sep:  "-",
			},
			want: "golang.png",
		},
		{
			name: "Does not include custom attachment names",
			args: args{
				path: "https://baike.seekhill.com/uploads/202106/1624354355xY7cLkuE_s.png",
			},
			want: "1624354355xY7cLkuE_s.png",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAttachmentName(tt.args.path, tt.args.sep); got != tt.want {
				t.Errorf("GetAttachmentName() = %v, want %v", got, tt.want)
			}
		})
	}
}
