package model

import "testing"

func TestPositionZero(t *testing.T) {
	var pos Position
	if pos.File != "" || pos.Line != 0 || pos.Column != 0 {
		t.Fatalf("zero Position = %+v", pos)
	}
}
