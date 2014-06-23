package quick

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestShrink(t *testing.T) {
}

func TestShrinkInt(t *testing.T) {
	tests := []struct{
		pred func(i int) bool
		fail, shrunk int
	}{
		//{func(i int) bool { return i < 14 }, 293, 14},
		{func(i int) bool { return -5 < i && i < 5 }, -99, 5},
	}

	for _, test := range tests {
		pred := func(v reflect.Value) bool { return test.pred(int(v.Int())) }
		shrunk := shrink(pred, reflect.ValueOf(test.fail)).Int()
		if int(shrunk) != test.shrunk {
			t.Errorf("expected %v, got %v", test.shrunk, shrunk)
		}
	}
}

func TestShrinkString(t *testing.T) {
	tests := []struct{
		pred func(s string) bool
		fail, shrunk string
	}{
		{func(s string) bool { return !strings.HasPrefix(s, "foo") }, "foobar", "foo"},
		{func(s string) bool { return utf8.RuneCountInString(s) == len(s) }, "Ħİ¡", "à"},
	}

	for _, test := range tests {
		pred := func(v reflect.Value) bool { return test.pred(v.String()) }
		shrunk := shrink(pred, reflect.ValueOf(test.fail)).String()
		if shrunk != test.shrunk {
			t.Errorf("expected %v, got %v", test.shrunk, shrunk)
		}
	}
}
