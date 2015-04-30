package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Scanner struct {
	x, y, units, xStep, yStep float64
	v, a                      float64
	relative, tubeOn          bool
}

var filter map[string]func(string) (byte, byte, bool)
var context Scanner

func setup() {
	filter = make(map[string]func(string) (byte, byte, bool))
	filter["X"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		if error != nil {
			panic("not a number")
		}
		tmp *= context.units * context.xStep
		if context.relative {
			context.x += tmp
		} else {
			context.x = tmp
		}
		return uint8(context.x / 256), uint8(context.x), true
	}
	filter["Y"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		if error != nil {
			panic("not a number")
		}
		tmp *= context.yStep
		if context.relative {
			context.y += tmp
		} else {
			context.y = tmp
		}
		return uint8(context.y / 256), uint8(context.y), true
	}
	filter["S"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseInt(str, 10, 16)
		if error != nil {
			panic("not a number")
		}
		return uint8(tmp / 256), uint8(tmp), true
	}
	filter["M"] = filter["S"]
	filter["V"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		context.v = tmp
		if error != nil {
			panic("not a number")
		}
		tmp *= 100
		return uint8(tmp / 256), uint8(tmp), true
	}
	filter["A"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		context.a = tmp
		if error != nil {
			panic("not a number")
		}
		tmp *= 1000
		return uint8(tmp / 256), uint8(tmp), true
	}

	context.xStep = 100
	context.yStep = 10
	context.units = 1
}
func main() {
	setup()

	fmt.Println(execute("G0 X1.4 Y1.2123")) //move to 1.4 1.2123
	fmt.Println(execute("G91"))             //set relative
	fmt.Println(execute("G0 X1.4"))         //move to 2.8 1.2123
}
func execute(line string) []byte {
	tokens := strings.Split(line, " ")
	switch tokens[0] {
	case "G0", "G1":
		return compile([]string{"Y", "X"}, 1, tokens, filter)
	case "G4":
		return compile([]string{"S", "M"}, 3, tokens, filter)
	case "G20":
		setScale(2.6)
	case "G21":
		setScale(1.0)
	case "G90":
		setRelative(false)
	case "G91":
		setRelative(true)
	case "M0":
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 5
		return tmp
	case "M1":
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 6
		return tmp
	case "M3":
		context.tubeOn = true
		return compile([]string{"V", "A"}, 7, tokens, filter)
	case "M5":
		context.tubeOn = false
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 8
		return tmp
	case "M100":
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 9
		return tmp
	case "M102":
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 0
		return tmp
	default:
		panic("Invalid command")
	}
	return nil
}

func setScale(scale float64) {
	context.units = scale
}
func setRelative(relative bool) {
	context.relative = relative
}

func compile(flags []string, code byte, tokens []string, filter map[string]func(string) (byte, byte, bool)) []byte {
	args := Parse(tokens, flags)
	out := make([]byte, 2)

	for i, arg := range args {
		first, second, set := filter[i](arg)
		if set {
			out = append(out, first, second)
		}
		out[1] += 1 << uint(IndexOf(i, flags))
	}

	out[0] = byte(len(out) - 1)

	out[1] = out[1]<<4 + code

	return out
}
func Parse(tokens []string, flags []string) map[string]string {
	out := make(map[string]string)
	for _, token := range tokens {
		for _, flag := range flags {
			if strings.HasPrefix(token, flag) {
				out[flag] = strings.TrimPrefix(token, flag)
			}
		}
	}
	return out
}
func IndexOf(elem string, array []string) int {
	for i, el := range array {
		if el == elem {
			return i
		}
	}
	return -1
}
