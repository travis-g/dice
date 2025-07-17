/*
Parser is an experiment to define grammar for dice roll expressions using a
parsing expression grammar (PEG) approach. The PEG is generated using pigeon:

	pigeon -o parser.go dice.peg

See https://github.com/mna/pigeon/
*/
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/travis-g/dice"
	"golang.org/x/exp/constraints"
)

var (
	ErrNilRollQuery = errors.New("nil RollQuery")
	ErrNilParameter = errors.New("parameter from context is nil")
)

type Label string

func (l *Label) String() string {
	if len(*l) == 0 {
		return ""
	}
	b := new(bytes.Buffer)
	b.WriteString("[")
	b.WriteString(string(*l))
	b.WriteString("]")
	return b.String()
}

type OpFactor struct {
	op     string
	factor interface{}
}

type Expression struct {
	l interface{}
	r []OpFactor
}

type Dice struct {
	die       string
	modifiers []string
}

type Operator int16

type Factor interface {
	// FIXME: cut out the any constraint
	constraints.Integer | constraints.Float | any
}

type Resolver[F Factor] interface {
	Resolve(ctx context.Context) error
	Value() F
}

type RollQuery struct {
	string
	answer interface{}
}

func (q *RollQuery) String() string {
	if q == nil {
		return "[<nil>]"
	}
	return fmt.Sprintf("[%s]", q.string)
}

func (q *RollQuery) Resolve(ctx context.Context) error {
	if q != nil {
		params, _ := dice.CtxParameters(ctx)
		if a, ok := params[q.string]; ok {
			q.answer = a
		} else {
			return ErrNilParameter
		}
	}
	return ErrNilRollQuery
}

func (q *RollQuery) Value() interface{} {
	if q == nil {
		return nil
	}
	return q.answer
}

var (
	_ = Resolver[any](&RollQuery{})
)

func anyLog(v any) {
	log.Println(v.(string))
}

func typeOf(v any) {
	log.Println(reflect.TypeOf(v), reflect.ValueOf(v))
}

func castAnySlice(v any) []any {
	if v == nil {
		return nil
	}
	return v.([]any)
}

func ptr(v any) *any {
	return &v
}

type Roller interface {
	Roll(context.Context) error
}

type Walker interface {
	Walk(context.Context) error
}

type Interface interface {
	Roller
	Walker
}

func main() {
	debug := flag.Bool("debug", false, "run parser in debug mode")
	flag.Parse()
	got, err := Parse("expression", []byte(flag.Arg(0)), Debug(*debug))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Printf("%#v", got)
	// b, _ := json.Marshal(got)
	// fmt.Printf("%s", b)
}
