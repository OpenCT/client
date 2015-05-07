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
	filter map[string]func(string) (byte, byte, bool)
}

func (ctx *Scanner) Setup() {
	ctx.filter = make(map[string]func(string) (byte, byte, bool))
	ctx.filter["X"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		if error != nil {
			panic("not a number")
		}
		tmp *= ctx.units * ctx.xStep
		if ctx.relative {
			ctx.x += tmp
		} else {
			ctx.x = tmp
		}
		return uint8(ctx.x / 256), uint8(ctx.x), true
	}
	ctx.filter["Y"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		if error != nil {
			panic("not a number")
		}
		tmp *= ctx.yStep
		if ctx.relative {
			ctx.y += tmp
		} else {
			ctx.y = tmp
		}
		return uint8(ctx.y / 256), uint8(ctx.y), true
	}
	ctx.filter["S"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseInt(str, 10, 16)
		if error != nil {
			panic("not a number")
		}
		return uint8(tmp / 256), uint8(tmp), true
	}
	ctx.filter["M"] = ctx.filter["S"]
	ctx.filter["V"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		ctx.v = tmp
		if error != nil {
			panic("not a number")
		}
		tmp *= 100
		return uint8(tmp / 256), uint8(tmp), true
	}
	ctx.filter["A"] = func(str string) (byte, byte, bool) {
		tmp, error := strconv.ParseFloat(str, 64)
		ctx.a = tmp
		if error != nil {
			panic("not a number")
		}
		tmp *= 1000
		return uint8(tmp / 256), uint8(tmp), true
	}

	ctx.xStep = 100
	ctx.yStep = 10
	ctx.units = 1
}
func main() {
  ctx := new(Scanner)
	ctx.Setup()
  
	fmt.Println(ctx.Execute("G0 X1.4 Y1.2123")) //move to 1.4 1.2123
	fmt.Println(ctx.Execute("G91"))             //set relative
	fmt.Println(ctx.Execute("G0 X1.4"))         //move to 2.8 1.2123
}
func (ctx *Scanner) Execute(line string) []byte { //[length,flags << 4 + code,args ...]
	tokens := strings.Split(line, " ")
	switch tokens[0] {
	case "G0", "G1": //Move to
		return compile([]string{"Y", "X"}, 1, tokens, ctx.filter)
	case "G4":  //Dwell
		return compile([]string{"S", "M"}, 3, tokens, ctx.filter)
	case "G20": //Unit = inches
		ctx.setScale(2.6)
	case "G21": //Unit = MM
		ctx.setScale(1.0)
	case "G90": //Absolute Positioning
		ctx.setRelative(false)
	case "G91": //Reletive Posititoning
		ctx.setRelative(true)
	case "M0":  //Stop
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 5
		return tmp
	case "M1": //Sleep
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 6
		return tmp
	case "M3": //Tube On
		ctx.tubeOn = true
		return compile([]string{"V", "A"}, 7, tokens, ctx.filter)
	case "M5": //Tube off
		ctx.tubeOn = false
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 8
		return tmp
	case "M100": //Grab data
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 9
		return tmp
	case "M102": //Version
		tmp := make([]byte, 2)
		tmp[0] = 1
		tmp[1] = 0
		return tmp
	default:
		panic("Invalid command")
	}
	return nil
}

func (ctx *Scanner) setScale(scale float64) {
	ctx.units = scale
}
func (ctx *Scanner) setRelative(relative bool) {
	ctx.relative = relative
}

func compile(flags []string, code byte, tokens []string, filter map[string]func(string) (byte, byte, bool)) []byte {
	args := parse(tokens, flags)
	out := make([]byte, 2)

	for i, arg := range args {
		first, second, set := filter[i](arg)
		if set {
			out = append(out, first, second)
		}
		out[1] += 1 << uint(indexOf(i, flags))
	}

	out[0] = byte(len(out) - 1)

	out[1] = out[1]<<4 + code

	return out
}
func parse(tokens []string, flags []string) map[string]string {
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
func indexOf(elem string, array []string) int {
	for i, el := range array {
		if el == elem {
			return i
		}
	}
	return -1
}
