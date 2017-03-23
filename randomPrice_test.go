package main

import "testing"

func Test_round(t *testing.T) {
	type args struct {
		f float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Round Test",
			args: args{f: 1.8},
			want: 1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := round(tt.args.f); got != tt.want {
				t.Errorf("round() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_random(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := random(tt.args.min, tt.args.max); got != tt.want {
				t.Errorf("random() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randomPrice_randomPrice(t *testing.T) {
	type fields struct {
		percentage float64
		price      float64
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &randomPrice{
				percentage: tt.fields.percentage,
				price:      tt.fields.price,
			}
			if got := rp.randomPrice(); got != tt.want {
				t.Errorf("randomPrice.randomPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}
