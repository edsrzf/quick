package quick

import (
	"reflect"
	"unicode"
	"unicode/utf8"
)

// A Shrinker can produce simpler values with its same type.
type Shrinker interface {
	// Shrink sends simpler values of its type to out. Shrink must stop
	// sending when it receives a value on stop. To avoid possible deadlocks,
	// all sends must be done in a select
	Shrink(out chan<- reflect.Value, stop <-chan bool)
}

// TryShrink is a convenience method for implementing Shrinker's Shrink method.
// It sends vals one-by-one to out until it has sent them all or received a value
// on stop.
// It returns true when a value was received on stop and the Shrinker should stop
// sending values.
func TryShrink(out chan<- reflect.Value, stop <-chan bool, vals ...interface{}) bool {
	for _, val := range vals {
		select {
		case out <- reflect.ValueOf(val):
		case <- stop:
			return true
		}
	}
	return false
}

func shrinkArguments(f func([]reflect.Value) bool, arguments []reflect.Value) []reflect.Value {
	newArgs := append([]reflect.Value(nil), arguments...)
	for i, arg := range arguments {
		pred := func(v reflect.Value) bool {
			newArgs[i] = v
			return f(newArgs)
		}
		newArgs[i] = shrink(pred, arg)
	}
	return newArgs
}

func shrinkInt(val reflect.Value, out chan<- reflect.Value, stop <-chan bool) {
	v := val.Int()
	if v == 0 {
		return
	}
	i := v / 2
	var done bool
	if v > 0 {
		done = TryShrink(out, stop, 0, v - i)
	} else {
		done = TryShrink(out, stop, 0, -v, v - i)
	}
	if done {
		return
	}
	for i != 0 {
		i /= 2
		if TryShrink(out, stop, v - i) {
			return
		}
	}
}

func shrinkUint(val reflect.Value, out chan<- reflect.Value, stop <-chan bool) {
	v := val.Int()
	if v == 0 {
		return
	}
	i := v / 2
	if TryShrink(out, stop, 0, v - i) {
		return
	}
	for i != 0 {
		i /= 2
		if TryShrink(out, stop, v - i) {
			return
		}
	}
}

func shrinkString(val reflect.Value, out chan<- reflect.Value, stop <-chan bool) {
	s := val.String()
	if TryShrink(out, stop, "") {
		return
	}
	// easier to work with runes
	runes := []rune(s)
	i := len(runes) / 2
	// first try to shrink the size
	for i > 0 {
		if TryShrink(out, stop, string(runes[i:]), string(runes[:i])) {
			return
		}
		i /= 2
	}
	// now convert to runes and shrink each individual rune
	newrunes := make([]rune, len(runes))
	for i, r := range runes {
		copy(newrunes, runes)
		if r >= utf8.RuneSelf {
			newrunes[i] = 'à'
			if TryShrink(out, stop, string(newrunes)) {
				return
			}
		}
		newrunes[i] = 'a'
		if TryShrink(out, stop, string(newrunes)) {
			return
		}
		if lower := unicode.ToLower(r); lower != r {
			newrunes[i] = lower
			if TryShrink(out, stop, string(newrunes)) {
				return
			}
		}
	}
}

func shrink(pred func(reflect.Value) bool, val reflect.Value) reflect.Value {
	for {
		if newVal := shrinkStep(pred, val); newVal.IsValid() && !reflect.DeepEqual(val.Interface(), newVal.Interface()) {
			val = newVal
		} else {
			return val
		}
	}
}

func shrinkStep(pred func(reflect.Value) bool, val reflect.Value) reflect.Value {
	out := make(chan reflect.Value)
	stop := make(chan bool, 1) // buffer size 1 in case we send after the last item
	go func() {
		defer close(out)
		iface := val.Interface()
		if shrinker, ok := iface.(Shrinker); ok {
			shrinker.Shrink(out, stop)
			return
		}
		switch v := iface.(type) {
		case bool:
			if v {
				out <- reflect.ValueOf(false)
			}
		case int, int8, int16, int32, int64:
			shrinkInt(val, out, stop)
		case uint, uint8, uint16, uint32, uint64, uintptr:
			shrinkUint(val, out, stop)
		case string:
			shrinkString(val, out, stop)
		}
	}()
	for v := range out {
		if !pred(v) {
			stop <- true
			return v
		}
	}
	return reflect.Value{}
}
