package main

import "testing"

func Test_toIntWithDefault(t *testing.T) {
	type args struct {
		val        string
		defaultVal int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "empty",
			args: args{val: "", defaultVal: 10},
			want: 10,
		},
		{
			name:    "invalid",
			args:    args{val: "invalid", defaultVal: 10},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{val: "20", defaultVal: 10},
			want: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toIntWithDefault(tt.args.val, tt.args.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("toIntWithDefault(%q, %d) error = %v, wantErr %v", tt.args.val, tt.args.defaultVal, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("toIntWithDefault(%q, %d) = %v, want %v", tt.args.val, tt.args.defaultVal, got, tt.want)
			}
		})
	}
}
