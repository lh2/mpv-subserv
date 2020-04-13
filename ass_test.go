package main

import "testing"

func TestAssText(t *testing.T) {
	pairs := map[string]string{
		"{}test{{}}": "test",
		"{\\fad(150,150)\\be35}{\\t(\\fscx115)}{\\org(-5000,436)\\frz-0.1\\t(0,250,\\frz0)\\t(350,850)}息づきしたら消えた": "息づきしたら消えた",
		"{\\fad(150,150)\\be35}{\\org(-5000,436)\\frz-0.1\\t(0,250,\\frz0)\\t(350,850)}あ　あれってハクビシン":               "あ　あれってハクビシン",
		"{\\an8}{\\fn方正中倩简体}{\\fs40}くれよ":                                                                          "くれよ",
		"私の名前は一条蛍": "私の名前は一条蛍",
	}
	for d, e := range pairs {
		a := parseAssText(d)
		if a != e {
			t.Errorf("actual=%s, expected=%s", a, e)
		}
	}
}

func TestAssTime(t *testing.T) {
	pairs := map[string]float64{
		"0:00:23.45": 23.45,
		"0:01:04":    64,
		"0:27:34.12": 1654.12,
		"1:23:45":    5025,
	}
	for d, e := range pairs {
		a := parseAssTime(d)
		if a != e {
			t.Errorf("actual=%v, expected=%v", a, e)
		}
	}
}
