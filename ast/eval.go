package main

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/travis-g/dice"
)

// Evaluate computes the numeric result of the AST rooted at a Root node.
// Query values are resolved from context parameters or query options, and dice
// notation is rolled using the dice package.
func (r *Root) Evaluate(ctx context.Context) (float64, error) {
	if r == nil {
		return 0, ErrInvalidStructField
	}
	if r.Expr == nil {
		return 0, nil
	}
	return r.Expr.Evaluate(ctx)
}

func (r *Root) Resolve(ctx context.Context) (Node, error) {
	if r == nil {
		return nil, nil
	}
	if r.Expr != nil {
		if _, err := r.Expr.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (e *Expr) Evaluate(ctx context.Context) (float64, error) {
	if e == nil || e.L == nil {
		return 0, ErrInvalidStructField
	}
	values := make([]float64, 0, len(e.R)+1)
	ops := make([]string, 0, len(e.R))

	first, err := e.L.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	values = append(values, first)

	for _, opTerm := range e.R {
		if opTerm == nil || opTerm.Term == nil {
			return 0, ErrInvalidStructField
		}
		operand, err := opTerm.Term.Evaluate(ctx)
		if err != nil {
			return 0, err
		}
		values = append(values, operand)
		ops = append(ops, opTerm.Op)
	}

	return evaluateExpressionValues(values, ops)
}

func (e *Expr) Resolve(ctx context.Context) (Node, error) {
	if e == nil {
		return nil, nil
	}
	if e.L != nil {
		if _, err := e.L.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	for _, opTerm := range e.R {
		if opTerm == nil {
			continue
		}
		if _, err := opTerm.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (f *Func) Evaluate(ctx context.Context) (float64, error) {
	if f == nil {
		return 0, ErrInvalidStructField
	}
	args := make([]float64, 0)
	if f.Args != nil {
		for _, arg := range f.Args.Arg {
			if arg == nil {
				return 0, ErrInvalidStructField
			}
			value, err := arg.Evaluate(ctx)
			if err != nil {
				return 0, err
			}
			args = append(args, value)
		}
	}
	return evaluateFunction(strings.ToLower(f.Name), args)
}

func (f *Func) Resolve(ctx context.Context) (Node, error) {
	if f == nil || f.Args == nil {
		return f, nil
	}
	return f.Args.Resolve(ctx)
}

func (q *Query) Evaluate(ctx context.Context) (float64, error) {
	if q == nil {
		return 0, ErrInvalidStructField
	}

	if value, ok := resolveQueryParameter(ctx, q.Name); ok {
		return evaluateExpressionString(ctx, value)
	}

	if len(q.Options) == 0 {
		return 0, fmt.Errorf("unresolved query %q", q.Name)
	}

	option := q.Options[0]
	value := strings.TrimSpace(option.OptionValue)
	if value == "" {
		value = strings.TrimSpace(option.OptionLabel)
	}
	if value == "" {
		return 0, fmt.Errorf("query %q has no resolvable option", q.Name)
	}
	return evaluateExpressionString(ctx, value)
}

func (q *Query) Resolve(ctx context.Context) (Node, error) {
	if q == nil {
		return nil, nil
	}
	if _, ok := resolveQueryParameter(ctx, q.Name); ok {
		return q, nil
	}
	if len(q.Options) == 0 {
		return q, fmt.Errorf("unresolved query %q", q.Name)
	}
	return q, nil
}

func (qo *QueryOption) Resolve(ctx context.Context) (Node, error) {
	return qo, nil
}

func (a *Args) Evaluate(ctx context.Context) (float64, error) {
	if a == nil || len(a.Arg) == 0 {
		return 0, ErrInvalidStructField
	}
	return a.Arg[0].Evaluate(ctx)
}

func (a *Args) Resolve(ctx context.Context) (Node, error) {
	if a == nil {
		return nil, nil
	}
	for _, arg := range a.Arg {
		if arg == nil {
			continue
		}
		if _, err := arg.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (ot *OpTerm) Resolve(ctx context.Context) (Node, error) {
	if ot == nil || ot.Term == nil {
		return ot, nil
	}
	_, err := ot.Term.Resolve(ctx)
	return ot, err
}

func (t *Term) Evaluate(ctx context.Context) (float64, error) {
	if t == nil {
		return 0, ErrInvalidStructField
	}
	switch {
	case t.Subexpr != nil:
		return t.Subexpr.Evaluate(ctx)
	case t.Query != nil:
		return t.Query.Evaluate(ctx)
	case t.Number != nil:
		return *t.Number, nil
	case t.Dice != nil:
		var notation strings.Builder
		notation.WriteString(t.Dice.String())
		for _, m := range t.Modifiers {
			notation.WriteString(m.String())
		}
		for _, gm := range t.GroupModifier {
			notation.WriteString(gm.String())
		}
		props, err := dice.ParseNotation(ctx, notation.String())
		if err != nil {
			return 0, err
		}
		group, err := dice.NewRollerGroup(&props)
		if err != nil {
			return 0, err
		}
		if err := group.FullRoll(ctx); err != nil {
			return 0, err
		}
		return group.Total(ctx)
	case t.Func != nil:
		return t.Func.Evaluate(ctx)
	default:
		return 0, ErrInvalidStructField
	}
}

func (t *Term) Resolve(ctx context.Context) (Node, error) {
	if t == nil {
		return nil, nil
	}
	if t.Subexpr != nil {
		if _, err := t.Subexpr.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	if t.Query != nil {
		if _, err := t.Query.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	if t.Func != nil {
		if _, err := t.Func.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (d *Dice) Resolve(ctx context.Context) (Node, error) {
	return d, nil
}

func (m *Modifier) Resolve(ctx context.Context) (Node, error) {
	return m, nil
}

func (gm *GroupModifier) Resolve(ctx context.Context) (Node, error) {
	return gm, nil
}

func evaluateExpressionValues(values []float64, ops []string) (float64, error) {
	if len(values) == 0 {
		return 0, ErrInvalidStructField
	}
	if len(values) != len(ops)+1 {
		return 0, ErrInvalidStructField
	}

	highPrecedence := map[string]bool{"**": true, "^": true}
	midPrecedence := map[string]bool{"*": true, "/": true, "%": true}
	lowPrecedence := map[string]bool{"+": true, "-": true}
	shiftPrecedence := map[string]bool{"<<": true, ">>": true}

	// Exponentiation
	for i := len(ops) - 1; i >= 0; i-- {
		if highPrecedence[ops[i]] {
			result, err := evaluateOperator(values[i], ops[i], values[i+1])
			if err != nil {
				return 0, err
			}
			values[i] = result
			values = append(values[:i+1], values[i+2:]...)
			ops = append(ops[:i], ops[i+1:]...)
		}
	}

	// Multiplication, division, modulo
	for i := 0; i < len(ops); {
		if midPrecedence[ops[i]] {
			result, err := evaluateOperator(values[i], ops[i], values[i+1])
			if err != nil {
				return 0, err
			}
			values[i] = result
			values = append(values[:i+1], values[i+2:]...)
			ops = append(ops[:i], ops[i+1:]...)
			continue
		}
		i++
	}

	// Addition, subtraction
	for i := 0; i < len(ops); {
		if lowPrecedence[ops[i]] {
			result, err := evaluateOperator(values[i], ops[i], values[i+1])
			if err != nil {
				return 0, err
			}
			values[i] = result
			values = append(values[:i+1], values[i+2:]...)
			ops = append(ops[:i], ops[i+1:]...)
			continue
		}
		i++
	}

	// Bit shifts
	for i := 0; i < len(ops); {
		if shiftPrecedence[ops[i]] {
			result, err := evaluateOperator(values[i], ops[i], values[i+1])
			if err != nil {
				return 0, err
			}
			values[i] = result
			values = append(values[:i+1], values[i+2:]...)
			ops = append(ops[:i], ops[i+1:]...)
			continue
		}
		i++
	}

	if len(values) != 1 {
		return 0, ErrInvalidStructField
	}
	return values[0], nil
}

func evaluateOperator(left float64, op string, right float64) (float64, error) {
	switch op {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, ErrDivisionByZero
		}
		return left / right, nil
	case "%":
		if right == 0 {
			return 0, ErrDivisionByZero
		}
		return math.Mod(left, right), nil
	case "**", "^":
		return math.Pow(left, right), nil
	case "<<":
		return float64(int64(left) << uint64(right)), nil
	case ">>":
		return float64(int64(left) >> uint64(right)), nil
	default:
		return 0, fmt.Errorf("unknown operator %q", op)
	}
}

func evaluateFunction(name string, args []float64) (float64, error) {
	switch name {
	case "abs":
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Abs(args[0]), nil
	case "ceil":
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Ceil(args[0]), nil
	case "floor":
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Floor(args[0]), nil
	case "max":
		if len(args) < 1 {
			return 0, ErrNotEnoughArgs
		}
		return slices.Max(args), nil
	case "min":
		if len(args) < 1 {
			return 0, ErrNotEnoughArgs
		}
		return slices.Min(args), nil
	case "round":
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Round(args[0]), nil
	default:
		return 0, fmt.Errorf("unknown function %q", name)
	}
}

func resolveQueryParameter(ctx context.Context, name string) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if params := ctx.Value(dice.CtxKeyParameters); params != nil {
		p, ok := params.(map[string]string)
		if !ok {
			return "", false
		}

		for k, v := range p {
			if strings.EqualFold(k, name) {
				return v, true
			}
		}
	}
	return "", false
}

func evaluateExpressionString(ctx context.Context, expression string) (float64, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return 0, ErrInvalidStructField
	}
	root, err := ParseString(expression, false)
	if err == nil {
		return root.Evaluate(ctx)
	}
	if value, err2 := strconv.ParseFloat(expression, 64); err2 == nil {
		return value, nil
	}
	return 0, fmt.Errorf("unable to evaluate expression %q: %w", expression, err)
}
