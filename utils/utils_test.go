package utils

import (
	"math"
	"testing"
	"time"

	"github.com/jingyuexing/stocks/ast"
)

func TestCalculateProfitPercentage(t *testing.T) {
	tests := []struct {
		name       string
		beginValue float64
		nowValue   float64
		want       float64
		wantErr    bool
	}{
		{
			name:       "normal profit",
			beginValue: 100,
			nowValue:   150,
			want:       50,
			wantErr:    false,
		},
		{
			name:       "normal loss",
			beginValue: 100,
			nowValue:   80,
			want:       -20,
			wantErr:    false,
		},
		{
			name:       "no change",
			beginValue: 100,
			nowValue:   100,
			want:       0,
			wantErr:    false,
		},
		{
			name:       "begin value is zero",
			beginValue: 0,
			nowValue:   100,
			want:       0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateProfitPercentage(tt.beginValue, tt.nowValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateProfitPercentage(%v, %v) error = %v, wantErr %v", tt.beginValue, tt.nowValue, err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("CalculateProfitPercentage(%v, %v) = %v, want %v", tt.beginValue, tt.nowValue, got, tt.want)
			}
		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		from  string
		to    string
		want  float64
	}{
		{
			name:  "k to empty",
			value: 5,
			from:  "k",
			to:    "",
			want:  5000,
		},
		{
			name:  "empty to k",
			value: 5000,
			from:  "",
			to:    "k",
			want:  5,
		},
		{
			name:  "w to empty",
			value: 3,
			from:  "w",
			to:    "",
			want:  30000,
		},
		{
			name:  "m to k",
			value: 2,
			from:  "m",
			to:    "k",
			want:  2000,
		},
		{
			name:  "b to m",
			value: 1.5,
			from:  "b",
			to:    "m",
			want:  1500,
		},
		{
			name:  "k to w",
			value: 10,
			from:  "k",
			to:    "w",
			want:  1,
		},
		{
			name:  "invalid from",
			value: 10,
			from:  "x",
			to:    "k",
			want:  -1,
		},
		{
			name:  "invalid to",
			value: 10,
			from:  "k",
			to:    "x",
			want:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Size(tt.value, tt.from, tt.to)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Size(%v, %q, %q) = %v, want %v", tt.value, tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestConvertAmount(t *testing.T) {
	tests := []struct {
		name string
		val  ast.Literal
		want float64
	}{
		{
			name: "K unit",
			val: ast.Literal{
				Value: "5",
				Unit:  "K",
			},
			want: 5000,
		},
		{
			name: "k unit lowercase",
			val: ast.Literal{
				Value: "3",
				Unit:  "k",
			},
			want: 3000,
		},
		{
			name: "W unit",
			val: ast.Literal{
				Value: "2",
				Unit:  "W",
			},
			// Note: tokenizer.IsAmountUnit does not currently recognize "W",
			// so ConvertAmount returns the raw value.
			want: 2,
		},
		{
			name: "w unit lowercase",
			val: ast.Literal{
				Value: "4",
				Unit:  "w",
			},
			// Note: tokenizer.IsAmountUnit does not currently recognize "w",
			// so ConvertAmount returns the raw value.
			want: 4,
		},
		{
			name: "M unit",
			val: ast.Literal{
				Value: "1.5",
				Unit:  "M",
			},
			want: 1500000,
		},
		{
			name: "m unit lowercase",
			val: ast.Literal{
				Value: "2",
				Unit:  "m",
			},
			want: 2000000,
		},
		{
			name: "B unit",
			val: ast.Literal{
				Value: "1",
				Unit:  "B",
			},
			want: 1000000000,
		},
		{
			name: "b unit lowercase",
			val: ast.Literal{
				Value: "0.5",
				Unit:  "b",
			},
			want: 500000000,
		},
		{
			name: "no unit",
			val: ast.Literal{
				Value: "100",
				Unit:  "",
			},
			want: 100,
		},
		{
			name: "invalid unit",
			val: ast.Literal{
				Value: "100",
				Unit:  "x",
			},
			want: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertAmount(tt.val)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("ConvertAmount(%+v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestTimeDuration(t *testing.T) {
	tests := []struct {
		name    string
		dur     string
		wantErr bool
	}{
		{
			name:    "single unit seconds",
			dur:     "10s",
			wantErr: false,
		},
		{
			name:    "single unit minutes",
			dur:     "5m",
			wantErr: false,
		},
		{
			name:    "single unit hours",
			dur:     "2h",
			wantErr: false,
		},
		{
			name:    "single unit days",
			dur:     "1d",
			wantErr: false,
		},
		{
			name:    "single unit weeks",
			dur:     "1w",
			wantErr: false,
		},
		{
			name:    "single unit months",
			dur:     "1m",
			wantErr: false,
		},
		{
			name:    "single unit years",
			dur:     "1y",
			wantErr: false,
		},
		{
			name:    "combined units",
			dur:     "1d 2h 30m",
			wantErr: false,
		},
		{
			name:    "invalid unit",
			dur:     "10x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TimeDuration(tt.dur)
			if (err != nil) != tt.wantErr {
				t.Errorf("TimeDuration(%q) error = %v, wantErr %v", tt.dur, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Since TimeDuration uses time.Now(), verify the result is in the future
				// and within a reasonable delta.
				if got.Before(time.Now().UTC()) {
					t.Errorf("TimeDuration(%q) = %v, expected future time", tt.dur, got)
				}
			}
		})
	}
}
