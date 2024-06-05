package main

import "testing"

func Test_toIntWithDefault(t *testing.T) {
	type args struct {
		value    string
		defaults int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "empty",
			args: args{value: "", defaults: 10},
			want: 10,
		},
		{
			name:    "invalid",
			args:    args{value: "invalid", defaults: 10},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{value: "20", defaults: 10},
			want: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := toIntWithDefault(tt.args.value, tt.args.defaults); (err != nil) != tt.wantErr {
				t.Errorf("toIntWithDefault() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if got != tt.want {
					t.Errorf("toIntWithDefault() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
