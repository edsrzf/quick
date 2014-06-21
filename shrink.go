package quick

import (
	"reflect"
)

type Shrinker interface {
	Shrink(in <-chan reflect.Value, out chan<- reflect.Value)
}

func shrinkArguments(f reflect.Value, arguments []reflect.Value) {
	args2 := append([]reflect.Value(nil), arguments...)
	for i, arg := range arguments {
		possibilities := shrink(arg)
	}
}

func shrinkInt(val reflect.Value, in <-chan reflect.Value, out chan<- reflect.Value) {
	out <- reflect.ValueOf(0).Convert(val.Type())
}

func shrinkUint(val reflect.Value, in <-chan reflect.Value, out chan<- reflect.Value) {
	v := val.Uint()
	typ := val.Type()
	out <- reflect.ValueOf(0).Convert(typ)
	i := v / 2
	for {
		if i == 0 {
			return
		}
		nextVal := reflect.ValueOf(v - i).Convert(typ)
		i /= 2
		select {
		case val = <-in:
			v = val.Uint()
			i = v / 2
		case out <- nextVal:
		}
	}
}

func shrinkString(val reflect.Value, in <-chan reflect.Value, out chan<- reflect.Value) {
	s := val.String()
	typ := val.Type()
	out <- reflect.ValueOf("").Convert(typ)
	i := len(s) / 2
	for {
		if i == len(s) {
			return
		}
		nextVal := reflect.ValueOf(s[:i]).Convert(typ)
		i /= 2
		select {
		case val = <-in:
			v = val.String()
			if v == "" {
				return
			}
			i = len(s) / 2
		case out <- nextVal:
		}
	}
}

func shrink(pred func(reflect.Value) bool, val reflect.Value) {
	in := make(chan reflect.Value)
	out := make(chan reflect.Value)
	defer close(out)

	go func() {
		if shrinker, ok := val.Interface().(Shrinker); ok {
			shrinker.Shrink(in, out)
			return
		}
		switch val.Kind() {
		case reflect.Bool:
			if val.Bool() {
				out <- reflect.ValueOf(false).Convert(val.Type())
			}
			return
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return shrinkInt(val)
		case reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return shrinkUint(val)
		}
	}()

	for {
		select {
		}
	}
}
