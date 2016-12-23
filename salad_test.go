package main

import "testing"

func TestSalad(t *testing.T) {
	text := `吾輩は猫である
今日は良い天気ですね
猫が寝転んだ
人間はひとくきの葦にすぎない`
	want := "今日はひとくきの葦にすぎない"
	h := NewSalad(text)
	result := h.makeWord(42)
	if result != want {
		t.Errorf("%s != %s", result, want)
	}
}
