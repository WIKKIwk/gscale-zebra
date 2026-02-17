package commands

import (
	"errors"
	"testing"
)

func TestIsInlineButtonUnsupported(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "button invalid", err: errors.New("telegram: Bad Request: BUTTON_TYPE_INVALID"), want: true},
		{name: "inline mode", err: errors.New("telegram: inline mode disabled"), want: true},
		{name: "other", err: errors.New("telegram: chat not found"), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isInlineButtonUnsupported(tc.err); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
