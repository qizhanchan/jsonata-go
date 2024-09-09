// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"fmt"
	"math"
	"reflect"
	"sort"

	"github.com/blues/jsonata-go/jlib"
	"github.com/blues/jsonata-go/jparse"
	"github.com/blues/jsonata-go/jtypes"
	"github.com/blues/jsonata-go/utils"
)

var undefined reflect.Value // TODO: 这里是否可以用其它方式代替

var typeInterfaceSlice = reflect.SliceOf(jtypes.TypeInterface)

// 这个函数只是一个路由，根据不同的 node 类型，调用不同的 eval 函数，本身没有其它功能
func eval(node jparse.Node, input reflect.Value, env *environment) (reflect.Value, error) {
	var err error
	var v reflect.Value

	switch node := node.(type) {
	case *jparse.StringNode:
		v, err = evalString(node, input, env)
	case *jparse.NumberNode:
		v, err = evalNumber(node, input, env)
	case *jparse.BooleanNode:
		v, err = evalBoolean(node, input, env)
	case *jparse.NullNode:
		v, err = evalNull(node, input, env)
	case *jparse.RegexNode:
		v, err = evalRegex(node, input, env)
	case *jparse.VariableNode:
		v, err = evalVariable(node, input, env)
	case *jparse.NameNode:
		v, err = evalName(node, input, env)
	case *jparse.PathNode:
		v, err = evalPath(node, input, env)
	case *jparse.NegationNode:
		v, err = evalNegation(node, input, env)
	case *jparse.RangeNode:
		v, err = evalRange(node, input, env)
	case *jparse.ArrayNode:
		v, err = evalArray(node, input, env)
	case *jparse.ObjectNode:
		v, err = evalObject(node, input, env)
	case *jparse.BlockNode:
		v, err = evalBlock(node, input, env)
	case *jparse.ConditionalNode:
		v, err = evalConditional(node, input, env)
	case *jparse.AssignmentNode:
		v, err = evalAssignment(node, input, env)
	case *jparse.WildcardNode:
		v, err = evalWildcard(node, input, env)
	case *jparse.DescendentNode:
		v, err = evalDescendent(node, input, env)
	case *jparse.GroupNode:
		v, err = evalGroup(node, input, env)
	case *jparse.PredicateNode:
		v, err = evalPredicate(node, input, env)
	case *jparse.SortNode:
		v, err = evalSort(node, input, env)
	case *jparse.LambdaNode:
		v, err = evalLambda(node, input, env)
	case *jparse.TypedLambdaNode:
		v, err = evalTypedLambda(node, input, env)
	case *jparse.ObjectTransformationNode:
		v, err = evalObjectTransformation(node, input, env)
	case *jparse.PartialNode:
		v, err = evalPartial(node, input, env)
	case *jparse.FunctionCallNode:
		v, err = evalFunctionCall(node, input, env)
	case *jparse.FunctionApplicationNode:
		v, err = evalFunctionApplication(node, input, env)
	case *jparse.NumericOperatorNode:
		v, err = evalNumericOperator(node, input, env)
	case *jparse.ComparisonOperatorNode:
		v, err = evalComparisonOperator(node, input, env)
	case *jparse.BooleanOperatorNode:
		v, err = evalBooleanOperator(node, input, env)
	case *jparse.StringConcatenationNode:
		v, err = evalStringConcatenation(node, input, env)
	default:
		panicf("eval: unexpected node type %T", node)
	}

	if err != nil {
		return undefined, err
	}

	// check if return is err
	if v.IsValid() && v.CanInterface() {
		if resErr, ok := v.Interface().(error); ok {
			return undefined, resErr
		}
	}

	if seq, ok := asSequence(v); ok {
		v = seq.Value()
	}

	return v, nil
}

func evalString(node *jparse.StringNode, data reflect.Value, env *environment) (reflect.Value, error) {
	return reflect.ValueOf(node.Value), nil
}

func evalNumber(node *jparse.NumberNode, data reflect.Value, env *environment) (reflect.Value, error) {
	return reflect.ValueOf(node.Value), nil
}

func evalBoolean(node *jparse.BooleanNode, data reflect.Value, env *environment) (reflect.Value, error) {
	return reflect.ValueOf(node.Value), nil
}

var null *interface{}

func evalNull(node *jparse.NullNode, data reflect.Value, env *environment) (reflect.Value, error) {
	return reflect.ValueOf(null), nil
}

func evalRegex(node *jparse.RegexNode, data reflect.Value, env *environment) (reflect.Value, error) {
	return reflect.ValueOf(newRegexCallable(node.Value)), nil
}

func evalVariable(node *jparse.VariableNode, data reflect.Value, env *environment) (reflect.Value, error) {
	if node.Name == "" {
		return data, nil
	}
	return env.lookup(node.Name), nil
}

func evalName(node *jparse.NameNode, data reflect.Value, env *environment) (reflect.Value, error) {
	var err error
	var v reflect.Value

	data = jtypes.Resolve(data)

	switch {
	case jtypes.IsStruct(data):
		v = data.FieldByName(node.Value)
	case jtypes.IsMap(data):
		v = data.MapIndex(reflect.ValueOf(node.Value))
	case jtypes.IsArray(data):
		v, err = evalNameArray(node, data, env)
	default:
		return undefined, nil
	}

	return v, err
}

func evalNameArray(node *jparse.NameNode, data reflect.Value, env *environment) (reflect.Value, error) {
	// utils.Log("evalNameArray")
	n := data.Len()
	results := make([]interface{}, 0, n)

	for i := 0; i < n; i++ {

		v, err := evalName(node, data.Index(i), env)
		if err != nil {
			return undefined, err
		}

		if v.IsValid() && v.CanInterface() {
			// check if v is array, just extend results, else append v to results
			if jtypes.IsArray(v) {
				// utils.Log("extend result")
				results = append(results, v.Interface().([]interface{})...)
			} else {
				// utils.Log("append result", v.Kind(), utils.GetJsonIndent(v), utils.GetJsonIndent(results))
				results = append(results, v.Interface())
			}

		}
	}
	utils.Log("evalNameArray result", utils.GetJsonIndent(results), "\n")

	return reflect.ValueOf(results), nil
}

func evalPath(node *jparse.PathNode, input reflect.Value, env *environment) (reflect.Value, error) {
	if len(node.Steps) == 0 {
		return undefined, nil
	}

	var isVar bool
	switch step0 := node.Steps[0].(type) {
	case (*jparse.VariableNode):
		isVar = true
	case (*jparse.PredicateNode):
		_, isVar = step0.Expr.(*jparse.VariableNode)
	}

	output := input
	if isVar || !jtypes.IsArray(input) {
		output = reflect.MakeSlice(typeInterfaceSlice, 1, 1)
		if input.IsValid() {
			output.Index(0).Set(input)
		}
	}

	var err error
	lastIndex := len(node.Steps) - 1
	for i, step := range node.Steps {

		if step0, ok := step.(*jparse.ArrayNode); ok && i == 0 {
			output, err = eval(step0, output, env)
			utils.Log("tag6")
		} else {
			utils.Log("tag7")
			output, err = evalPathStep(step, output, env, i == lastIndex)
		}

		if err != nil || output == undefined {
			return undefined, err
		}

		if jtypes.IsArray(output) && jtypes.Resolve(output).Len() == 0 {
			return undefined, nil
		}
	}

	if node.KeepArrays {
		utils.Log("结果似乎要转成列表", "output", utils.GetJsonIndent(output), "\n")
		if seq, ok := asSequence(output); ok {
			utils.Log("单例")
			seq.keepSingletons = true
			return reflect.ValueOf(seq), nil
		} else {
			output2 := reflect.MakeSlice(typeInterfaceSlice, 1, 1)
			output2.Index(0).Set(output)
			output = output2
		}
	}

	return output, nil
}

func evalPathStep(step jparse.Node, input reflect.Value, env *environment, lastStep bool) (reflect.Value, error) {
	var err error
	var results []reflect.Value
	utils.Log("")
	utils.Log("evalPathStep start")
	// TODO: 不知道为什么要转成列表，似乎可以直接取值就行
	// 似乎只需要在最后做类型转换就行，无论是数组还是二维数组
	if seq, ok := asSequence(input); ok {
		utils.Log("tag1")
		results, err = evalOverSequence(step, seq, env)
	} else {
		utils.Log("tag2")
		results, err = evalOverArray(step, input, env)
	}
	utils.Log("evalPathStep", utils.GetJsonIndent(step), utils.GetJsonIndent(input), "results", utils.GetJsonIndent(results))

	if err != nil {
		return undefined, err
	}

	if lastStep && len(results) == 1 && jtypes.IsArray(results[0]) {
		utils.Log("tag3")
		return results[0], nil
	}

	_, isCons := step.(*jparse.ArrayNode)
	resultSequence := newSequence(len(results))

	for _, v := range results {

		if isCons || !jtypes.IsArray(v) {
			if v.CanInterface() {
				resultSequence.Append(v.Interface())
			}
			continue
		}

		v = arrayify(v)
		for i, N := 0, v.Len(); i < N; i++ {
			if vi := v.Index(i); vi.IsValid() && vi.CanInterface() {
				resultSequence.Append(vi.Interface())
			}
		}
	}

	utils.Log("resultSequence", utils.GetJsonIndent(resultSequence))

	if resultSequence.Len() == 0 {
		utils.Log("return undefined")
		utils.Log("evalPathStep end")
		utils.Log("")
		return reflect.ValueOf(jtypes.NoMatchedCtx), nil
	}
	utils.Log("evalPathStep end")
	utils.Log("")

	return reflect.ValueOf(resultSequence), nil
}

func evalOverArray(node jparse.Node, data reflect.Value, env *environment) ([]reflect.Value, error) {
	var results []reflect.Value

	// 当前的逻辑是，如果是数组，就遍历数组，如果是对象，就遍历对象的值
	// 对于数组的 PredicateNode, 是要先执行 Exprs, 再执行 Filters
	if pred, ok := node.(*jparse.PredicateNode); ok {
		items, err := eval(pred.Expr, data, env)
		if err != nil || items == undefined {
			return results, err
		}

		utils.Log("after expr eval", utils.GetJsonIndent(items), utils.GetJsonIndent(data))

		for _, filter := range pred.Filters {
			// TODO: If this filter is of type *jparse.NumberNode,

			utils.Log("before filter item", utils.GetJsonIndent(items))

			items, err = applyFilter(filter, arrayify(items), env)
			if err != nil {
				return results, err
			}
			utils.Log("after filter item", utils.GetJsonIndent(items))

			if items.Len() == 0 {
				items = undefined
				break
			}
		}

		res := normalizeArray(items)
		if res.IsValid() {
			if results == nil {
				results = make([]reflect.Value, 0, data.Len())
			}
			results = append(results, res)
		}
		return results, nil
	}

	for i, N := 0, data.Len(); i < N; i++ {
		utils.Log("evalOverArray", "i", i, utils.GetJsonIndent(node), utils.GetJsonIndent(data.Index(i)))
		res, err := eval(node, data.Index(i), env)
		if err != nil {
			return nil, err
		}

		if res.IsValid() {
			if results == nil {
				results = make([]reflect.Value, 0, N)
			}
			results = append(results, res)
		}
	}

	return results, nil
}

func evalOverSequence(node jparse.Node, seq *Sequence, env *environment) ([]reflect.Value, error) {
	var results []reflect.Value

	for i, N := 0, len(seq.Values); i < N; i++ {

		res, err := eval(node, reflect.ValueOf(seq.Values[i]), env)
		if err != nil {
			return nil, err
		}

		if res.IsValid() {
			if results == nil {
				results = make([]reflect.Value, 0, N)
			}
			results = append(results, res)
		}
	}

	return results, nil
}

func evalNegation(node *jparse.NegationNode, data reflect.Value, env *environment) (reflect.Value, error) {
	rhs, err := eval(node.RHS, data, env)
	if err != nil || rhs == undefined {
		return undefined, err
	}

	n, ok := jtypes.AsNumber(rhs)
	if !ok {
		return undefined, newEvalError(ErrNonNumberRHS, node.RHS, "-")
	}

	return reflect.ValueOf(-n), nil
}

// maxRangeItems is the maximum array size allowed in a range
// expression. It's defined as a global so we can use it in
// the tests.
// We use the maximum value allowed by the jsonata-js library
const maxRangeItems = 10000000

func isInteger(x float64) bool {
	return x == math.Trunc(x)
}

func evalRange(node *jparse.RangeNode, data reflect.Value, env *environment) (reflect.Value, error) {
	evaluate := func(node jparse.Node) (float64, bool, bool, error) {

		v, err := eval(node, data, env)
		if err != nil || v == undefined {
			return 0, false, false, err
		}

		n, isNum := jtypes.AsNumber(v)
		return n, true, isNum && isInteger(n), nil
	}

	// Evaluate both sides and return any errors.
	lhs, lhsOK, lhsInteger, err := evaluate(node.LHS)
	if err != nil {
		return undefined, err
	}

	rhs, rhsOK, rhsInteger, err := evaluate(node.RHS)
	if err != nil {
		return undefined, err
	}

	// If either side is not an integer, return an error.
	if lhsOK && !lhsInteger {
		return undefined, newEvalError(ErrNonIntegerLHS, node.LHS, "..")
	}

	if rhsOK && !rhsInteger {
		return undefined, newEvalError(ErrNonIntegerRHS, node.RHS, "..")
	}

	// If either side is undefined or the left side is greater
	// than the right, return undefined.
	if !lhsOK || !rhsOK || lhs > rhs {
		return undefined, nil
	}

	size := int(rhs-lhs) + 1
	// Check for integer overflow or an array size that exceeds
	// our upper bound.
	if size < 0 || size > maxRangeItems {
		return undefined, newEvalError(ErrMaxRangeItems, "..", nil)
	}

	results := reflect.MakeSlice(typeInterfaceSlice, size, size)

	for i := 0; i < size; i++ {
		results.Index(i).Set(reflect.ValueOf(lhs))
		lhs++
	}

	return results, nil
}

func evalArray(node *jparse.ArrayNode, data reflect.Value, env *environment) (reflect.Value, error) {
	// Create a slice with capacity equal to the number of items
	// in the ArrayNode. Note that the final length of the array
	// may differ because:
	//
	// 1. Items that evaluate to undefined are excluded, reducing
	//    the length of the array.
	//
	// 2. Items that evaluate to arrays may be flattened into their
	//    individual elements, increasing the length of the array.
	results := make([]interface{}, 0, len(node.Items))

	for _, item := range node.Items {

		v, err := eval(item, data, env)
		if err != nil {
			return undefined, err
		}

		if v == undefined {
			continue
		}

		switch item.(type) {
		case *jparse.ArrayNode:
			if v.CanInterface() {
				results = append(results, v.Interface())
			}
		default:
			v = arrayify(v)
			for i, N := 0, v.Len(); i < N; i++ {
				if vi := v.Index(i); vi.IsValid() && vi.CanInterface() {
					results = append(results, vi.Interface())
				}
			}
		}
	}

	return reflect.ValueOf(results), nil
}

func evalObject(node *jparse.ObjectNode, data reflect.Value, env *environment) (reflect.Value, error) {
	utils.Log("eval ObjectNode")
	data = makeArray(data)

	keys, err := groupItemsByKey(node, data, env)
	if err != nil {
		return undefined, err
	}

	utils.Log("keys", utils.GetJsonIndent(keys))
	nItems := data.Len()
	results := make(map[string]interface{}, len(keys))

	for key, idx := range keys {

		items := data
		if n := len(idx.items); n != 0 && n != nItems {
			items = reflect.MakeSlice(typeInterfaceSlice, n, n)
			for i, j := range idx.items {
				items.Index(i).Set(data.Index(j))
			}
		}
		utils.Log("items", idx.pair, utils.GetJsonIndent(items), utils.GetJsonIndent(node.Pairs[idx.pair][1]))

		value, err := eval(node.Pairs[idx.pair][1], items, env)
		if err != nil {
			return undefined, err
		}
		utils.Log("key", key, "value", utils.GetJsonIndent(value))

		if value.IsValid() && value.CanInterface() {
			results[key] = value.Interface()
		}
	}

	return reflect.ValueOf(results), nil
}

type keyIndexes struct {
	pair  int
	items []int
}

func groupItemsByKey(obj *jparse.ObjectNode, items reflect.Value, env *environment) (map[string]keyIndexes, error) {
	nItems := items.Len()
	results := make(map[string]keyIndexes, len(obj.Pairs))

	for i, pair := range obj.Pairs {

		keyNode := pair[0]

		if s, ok := keyNode.(*jparse.StringNode); ok {

			key := s.Value
			if _, ok := results[key]; ok {
				return nil, newEvalError(ErrDuplicateKey, keyNode, key)
			}

			results[key] = keyIndexes{
				pair: i,
			}
			continue
		}

		for j := 0; j < nItems; j++ {

			v, err := eval(keyNode, items.Index(j), env)
			if err != nil {
				return nil, err
			}

			key, ok := jtypes.AsString(v)
			if !ok {
				return nil, newEvalError(ErrIllegalKey, keyNode, nil)
			}

			idx, ok := results[key]
			if !ok {
				results[key] = keyIndexes{
					pair:  i,
					items: []int{j},
				}
				continue
			}

			if idx.pair != i {
				return nil, newEvalError(ErrDuplicateKey, keyNode, key)
			}

			idx.items = append(idx.items, j)
			results[key] = idx
		}
	}

	return results, nil
}

func evalBlock(node *jparse.BlockNode, data reflect.Value, env *environment) (reflect.Value, error) {
	utils.Log("eval BlockNode")
	var err error
	var res reflect.Value

	// Create a local environment. Any variables defined
	// inside the block will be scoped to the block.
	// TODO: Is it worth calculating how many variables
	// are defined in the block so that we can create an
	// environment of the correct size?
	env = newEnvironment(env, 0)

	// Evaluate all expressions in the block.
	for _, node := range node.Exprs {
		// print node type
		// fmt.Printf("%#+v", node)
		res, err = eval(node, data, env)
		// check if res is err
		if res.IsValid() && res.CanInterface() {
			if resErr, ok := res.Interface().(error); ok {
				return undefined, resErr
			}
		}
		utils.Log("eval BlockNode", res, res.IsValid(), res.CanInterface())
		if err != nil {
			return undefined, err
		}
	}

	// Return the result of the last expression.
	return res, nil
}

func evalConditional(node *jparse.ConditionalNode, data reflect.Value, env *environment) (reflect.Value, error) {
	v, err := eval(node.If, data, env)
	if err != nil {
		return undefined, err
	}

	if jlib.Boolean(v) {
		return eval(node.Then, data, env)
	}

	if node.Else != nil {
		return eval(node.Else, data, env)
	}

	return undefined, nil
}

func evalAssignment(node *jparse.AssignmentNode, data reflect.Value, env *environment) (reflect.Value, error) {
	v, err := eval(node.Value, data, env)
	if err != nil {
		return undefined, err
	}

	env.bind(node.Name, v)
	return v, nil
}

func evalWildcard(node *jparse.WildcardNode, data reflect.Value, env *environment) (reflect.Value, error) {
	results := newSequence(0)

	walkObjectValues(data, func(v reflect.Value) {
		appendWildcard(results, v)
	})

	return reflect.ValueOf(results), nil
}

func appendWildcard(seq *Sequence, v reflect.Value) {
	switch {
	case jtypes.IsArray(v):
		v = flattenArray(v)
		for i, N := 0, v.Len(); i < N; i++ {
			if vi := v.Index(i); vi.IsValid() && vi.CanInterface() {
				seq.Append(vi.Interface())
			}
		}
	default:
		if v.IsValid() && v.CanInterface() {
			seq.Append(v.Interface())
		}
	}
}

func evalDescendent(node *jparse.DescendentNode, data reflect.Value, env *environment) (reflect.Value, error) {
	results := newSequence(0)

	recurseDescendents(results, data)

	return reflect.ValueOf(results), nil
}

func recurseDescendents(seq *Sequence, v reflect.Value) {
	if v.IsValid() && v.CanInterface() && !jtypes.IsArray(v) {
		seq.Append(v.Interface())
	}

	walkObjectValues(v, func(v reflect.Value) {
		recurseDescendents(seq, v)
	})
}

func evalGroup(node *jparse.GroupNode, data reflect.Value, env *environment) (reflect.Value, error) {
	items, err := eval(node.Expr, data, env)
	if err != nil {
		return undefined, err
	}

	return evalObject(node.ObjectNode, items, env)
}

func evalPredicate(node *jparse.PredicateNode, data reflect.Value, env *environment) (reflect.Value, error) {
	items, err := eval(node.Expr, data, env)
	if err != nil || items == undefined {
		return undefined, err
	}

	utils.Log("origin before filter", utils.GetJsonIndent(items), utils.GetJsonIndent(data))

	for _, filter := range node.Filters {

		// TODO: If this filter is of type *jparse.NumberNode,
		// we should access the indexed item directly instead
		// of calling applyFilter.

		utils.Log("origin before filter item", utils.GetJsonIndent(items))

		items, err = applyFilter(filter, arrayify(items), env)
		if err != nil {
			return undefined, err
		}
		utils.Log("origin after filter item", utils.GetJsonIndent(items))

		if items.Len() == 0 {
			items = undefined
			break
		}
	}

	return normalizeArray(items), nil
}

func applyFilter(filter jparse.Node, items reflect.Value, env *environment) (reflect.Value, error) {
	nItems := items.Len()
	results := reflect.MakeSlice(typeInterfaceSlice, 0, 0)

	for i := 0; i < nItems; i++ {

		item := items.Index(i)

		res, err := eval(filter, item, env)
		if err != nil {
			return undefined, err
		}

		if jtypes.IsNumber(res) {
			res = arrayify(res)
		}

		switch {
		case jtypes.IsArrayOf(res, jtypes.IsNumber):
			for j, N := 0, res.Len(); j < N; j++ {

				n, _ := jtypes.AsNumber(res.Index(j))
				index := int(math.Floor(n))
				if index < 0 {
					index += nItems
				}

				if index == i {
					results = reflect.Append(results, item)
				}
			}
		case jlib.Boolean(res):
			results = reflect.Append(results, item)
		}
	}

	return results, nil
}

type sortinfo struct {
	index  int
	values []reflect.Value
}

func buildSortInfo(items reflect.Value, terms []jparse.SortTerm, env *environment) ([]*sortinfo, error) {
	info := make([]*sortinfo, items.Len())

	isNumberTerm := make([]bool, len(terms))
	isStringTerm := make([]bool, len(terms))

	for i, N := 0, items.Len(); i < N; i++ {

		item := items.Index(i)
		values := make([]reflect.Value, len(terms))

		for j, term := range terms {

			v, err := eval(term.Expr, item, env)
			if err != nil {
				return nil, err
			}

			if v == undefined {
				continue
			}

			switch {
			case jtypes.IsNumber(v):
				if isStringTerm[j] {
					return nil, newEvalError(ErrSortMismatch, term.Expr, nil)
				}
				values[j] = v
				isNumberTerm[j] = true

			case jtypes.IsString(v):
				if isNumberTerm[j] {
					return nil, newEvalError(ErrSortMismatch, term.Expr, nil)
				}
				values[j] = v
				isStringTerm[j] = true

			default:
				return nil, newEvalError(ErrNonSortable, term.Expr, nil)
			}
		}

		info[i] = &sortinfo{
			index:  i,
			values: values,
		}
	}

	return info, nil
}

func makeLessFunc(info []*sortinfo, terms []jparse.SortTerm) func(int, int) bool {
	return func(i, j int) bool {
	Loop:
		for t, term := range terms {

			vi := info[i].values[t]
			vj := info[j].values[t]

			switch {
			case vi == undefined && vj == undefined:
				continue Loop
			case vi == undefined:
				return false
			case vj == undefined:
				return true
			}

			if eq(vi, vj) {
				continue Loop
			}

			if term.Dir == jparse.SortDescending {
				return lt(vj, vi)
			}
			return lt(vi, vj)
		}

		return false
	}
}

func evalSort(node *jparse.SortNode, data reflect.Value, env *environment) (reflect.Value, error) {
	items, err := eval(node.Expr, data, env)
	if err != nil || items == undefined {
		return undefined, err
	}

	items = arrayify(items)

	info, err := buildSortInfo(items, node.Terms, env)
	if err != nil {
		return undefined, err
	}

	sort.SliceStable(info, makeLessFunc(info, node.Terms))

	results := reflect.MakeSlice(typeInterfaceSlice, len(info), len(info))

	for i := range info {
		results.Index(i).Set(items.Index(info[i].index))
	}

	return normalizeArray(results), nil
}

func evalLambda(node *jparse.LambdaNode, data reflect.Value, env *environment) (reflect.Value, error) {
	f := &lambdaCallable{
		callableName: callableName{
			name: "lambda",
		},
		paramNames: node.ParamNames,
		body:       node.Body,
		context:    data,
		env:        env,
	}

	return reflect.ValueOf(f), nil
}

func evalTypedLambda(node *jparse.TypedLambdaNode, data reflect.Value, env *environment) (reflect.Value, error) {
	f := &lambdaCallable{
		callableName: callableName{
			name: "lambda",
		},
		typed:      true,
		params:     node.In,
		paramNames: node.ParamNames,
		body:       node.Body,
		context:    data,
		env:        env,
	}

	return reflect.ValueOf(f), nil
}

func evalObjectTransformation(node *jparse.ObjectTransformationNode, data reflect.Value, env *environment) (reflect.Value, error) {
	f := &transformationCallable{
		callableName: callableName{
			"transform",
		},
		pattern: node.Pattern,
		updates: node.Updates,
		deletes: node.Deletes,
		env:     env,
	}

	return reflect.ValueOf(f), nil
}

func evalPartial(node *jparse.PartialNode, data reflect.Value, env *environment) (reflect.Value, error) {
	v, err := eval(node.Func, data, env)
	if err != nil {
		return undefined, err
	}

	fn, ok := jtypes.AsCallable(v)
	if !ok {
		return undefined, newEvalError(ErrNonCallablePartial, node.Func, nil)
	}

	f := &partialCallable{
		callableName: callableName{
			name: fn.Name() + "_partial",
		},
		fn:      fn,
		args:    node.Args,
		context: data,
		env:     env,
	}

	return reflect.ValueOf(f), nil
}

type nameSetter interface {
	SetName(string)
}

type contextSetter interface {
	SetContext(reflect.Value)
}

func evalFunctionCall(node *jparse.FunctionCallNode, data reflect.Value, env *environment) (reflect.Value, error) {
	v, err := eval(node.Func, data, env)
	if err != nil {
		return undefined, err
	}

	fn, ok := jtypes.AsCallable(v)
	utils.Log("")
	utils.Log("evalFunctionCall", ok, v, fn)
	if !ok {
		return undefined, newEvalError(ErrNonCallable, node.Func, nil)
	}

	if setter, ok := fn.(nameSetter); ok {
		if sym, ok := node.Func.(*jparse.VariableNode); ok {
			utils.Log("sym", sym.Name)
			setter.SetName(sym.Name)
		}
	}

	if setter, ok := fn.(contextSetter); ok {
		utils.Log("data", utils.GetJsonIndent(data))
		setter.SetContext(data)
	}

	argv := make([]reflect.Value, len(node.Args))
	for i, arg := range node.Args {

		v, err := eval(arg, data, env)
		if err != nil {
			return undefined, err
		}
		utils.Log(i, "arg", utils.GetJsonIndent(arg), "v", utils.GetJsonIndent(v))

		argv[i] = v
	}
	utils.Log("argv", utils.GetJsonIndent(argv))
	utils.Log("")

	return fn.Call(argv)
}

func evalFunctionApplication(node *jparse.FunctionApplicationNode, data reflect.Value, env *environment) (reflect.Value, error) {
	// If the right hand side is a function call, insert
	// the left hand side into the argument list and
	// evaluate it.
	if f, ok := node.RHS.(*jparse.FunctionCallNode); ok {

		f.Args = append([]jparse.Node{node.LHS}, f.Args...)
		return evalFunctionCall(f, data, env)
	}

	// Evaluate both sides and return any errors.
	lhs, err := eval(node.LHS, data, env)
	if err != nil {
		return undefined, err
	}

	rhs, err := eval(node.RHS, data, env)
	if err != nil {
		return undefined, err
	}

	// Check that the right hand side is callable.
	f2, ok := jtypes.AsCallable(rhs)
	if !ok {
		return undefined, newEvalError(ErrNonCallableApply, node.RHS, "~>")
	}

	// If the left hand side is not callable, call the right
	// hand side using the left hand side as the argument.
	if !jtypes.IsCallable(lhs) {
		return f2.Call([]reflect.Value{lhs})
	}

	// Otherwise, combine both sides into a single callable.
	f1, _ := jtypes.AsCallable(lhs)

	f := &chainCallable{
		callables: []jtypes.Callable{
			f1,
			f2,
		},
	}

	return reflect.ValueOf(f), nil
}

func evalNumericOperator(node *jparse.NumericOperatorNode, data reflect.Value, env *environment) (reflect.Value, error) {
	evaluate := func(node jparse.Node) (float64, bool, bool, error) {

		v, err := eval(node, data, env)
		if err != nil || v == undefined {
			return 0, false, false, err
		}
		n, isNum := jtypes.AsNumber(v)
		return n, true, isNum, nil
	}

	// Evaluate both sides and return any errors.
	lhs, lhsOK, lhsNumber, err := evaluate(node.LHS)
	if err != nil {
		return undefined, err
	}

	rhs, rhsOK, rhsNumber, err := evaluate(node.RHS)
	if err != nil {
		return undefined, err
	}

	// Return an error if either side is not a number.
	if lhsOK && !lhsNumber {
		utils.Log("lhs", utils.GetJsonIndent(node.LHS), lhs, lhsOK, lhsNumber)
		utils.Log("rhs", utils.GetJsonIndent(node.RHS), rhs, rhsOK, rhsNumber)
		return undefined, newEvalError(ErrNonNumberLHS, node.LHS, node.Type)
	}

	if rhsOK && !rhsNumber {
		return undefined, newEvalError(ErrNonNumberRHS, node.RHS, node.Type)
	}

	// Return undefined if either side is undefined.
	if !lhsOK || !rhsOK {
		return undefined, nil
	}

	var x float64

	switch node.Type {
	case jparse.NumericAdd:
		x = lhs + rhs
	case jparse.NumericSubtract:
		x = lhs - rhs
	case jparse.NumericMultiply:
		x = lhs * rhs
	case jparse.NumericDivide:
		x = lhs / rhs
	case jparse.NumericModulo:
		x = math.Mod(lhs, rhs)
	default:
		panicf("unrecognised numeric operator %q", node.Type)
	}

	if math.IsInf(x, 0) {
		return undefined, newEvalError(ErrNumberInf, nil, node.Type)
	}

	if math.IsNaN(x) {
		return undefined, newEvalError(ErrNumberNaN, nil, node.Type)
	}

	return reflect.ValueOf(x), nil
}

// See https://docs.jsonata.org/expressions#comparison-expressions
func evalComparisonOperator(node *jparse.ComparisonOperatorNode, data reflect.Value, env *environment) (reflect.Value, error) {
	evaluate := func(node jparse.Node) (reflect.Value, bool, bool, error) {

		v, err := eval(node, data, env)
		if err != nil || v == undefined {
			utils.Log("got undefined")
			return undefined, false, false, err
		}

		return v, jtypes.IsNumber(v), jtypes.IsString(v), nil

	}

	utils.Log("")
	utils.Log("evalComparisonOperator start")
	utils.Log("node type", node.Type)

	// Evaluate both sides and return any errors.
	lhs, lhsNumber, lhsString, err := evaluate(node.LHS)
	if err != nil {
		return undefined, err
	}

	rhs, rhsNumber, rhsString, err := evaluate(node.RHS)
	if err != nil {
		return undefined, err
	}

	// If this operator requires comparable types, return
	// an error if a) either side is not comparable or b)
	// left side type does not equal right side type.
	if needComparableTypes(node.Type) {
		if lhs != undefined && !lhsNumber && !lhsString {
			return undefined, newEvalError(ErrNonComparableLHS, node.LHS, node.Type)
		}

		if rhs != undefined && !rhsNumber && !rhsString {
			return undefined, newEvalError(ErrNonComparableRHS, node.RHS, node.Type)
		}

		if lhs != undefined && rhs != undefined &&
			(lhsNumber != rhsNumber || lhsString != rhsString) {
			return undefined, newEvalError(ErrTypeMismatch, nil, node.Type)
		}
	}

	// Return undefined if either side is undefined.
	if lhs == undefined || rhs == undefined {
		utils.Log("lhs or rhs is undefined")
		return undefined, nil
	}

	var b bool

	switch node.Type {
	case jparse.ComparisonIn:
		b = in(lhs, rhs)
	case jparse.ComparisonEqual:
		b = eq(lhs, rhs)
	case jparse.ComparisonNotEqual:
		b = !eq(lhs, rhs)
	case jparse.ComparisonLess:
		b = lt(lhs, rhs)
	case jparse.ComparisonLessEqual:
		b = lte(lhs, rhs)
	case jparse.ComparisonGreater:
		b = !lte(lhs, rhs)
	case jparse.ComparisonGreaterEqual:
		b = !lt(lhs, rhs)
	default:
		panicf("unrecognised comparison operator %q", node.Type)
	}
	utils.Log("evalComparisonOperator end")

	return reflect.ValueOf(b), nil
}

func needComparableTypes(op jparse.ComparisonOperator) bool {
	switch op {
	case jparse.ComparisonEqual, jparse.ComparisonNotEqual, jparse.ComparisonIn:
		return false
	default:
		return true
	}
}

func eq(lhs, rhs reflect.Value) bool {
	// Numbers, strings, arrays, objects and booleans are compared by value.
	// Two strings might be different objects in memory but
	// they're still considered equal if they have the
	// same value.

	if v1, ok := jtypes.AsNumber(lhs); ok {
		v2, ok := jtypes.AsNumber(rhs)
		return ok && v1 == v2
	}

	if v1, ok := jtypes.AsString(lhs); ok {
		v2, ok := jtypes.AsString(rhs)
		return ok && v1 == v2
	}

	if v1, ok := jtypes.AsBool(lhs); ok {
		v2, ok := jtypes.AsBool(rhs)
		return ok && v1 == v2
	}

	// Arrays and maps are compared with a deep equal
	if jtypes.IsArray(lhs) && jtypes.IsArray(rhs) {
		return reflect.DeepEqual(lhs.Interface(), rhs.Interface())
	}

	if jtypes.IsMap(lhs) && jtypes.IsMap(rhs) {
		return reflect.DeepEqual(lhs.Interface(), rhs.Interface())
	}

	// All other types (e.g. functions) are
	// compared directly. Two functions with the same contents
	// are not considered equal unless they're the same
	// physical object in memory.

	return lhs == rhs
}

func lt(lhs, rhs reflect.Value) bool {
	if v1, ok := jtypes.AsNumber(lhs); ok {
		if v2, ok := jtypes.AsNumber(rhs); ok {
			return v1 < v2
		}
	}

	if v1, ok := jtypes.AsString(lhs); ok {
		if v2, ok := jtypes.AsString(rhs); ok {
			return v1 < v2
		}
	}

	panicf("lt: invalid types: lhs %s, rhs %s", lhs.Kind(), rhs.Kind())
	return false
}

func lte(lhs, rhs reflect.Value) bool {
	return lt(lhs, rhs) || eq(lhs, rhs)
}

func in(lhs, rhs reflect.Value) bool {
	// TODO: Does not work with null, e.g.
	//    null in null    // evaluates to false
	//    null in [null]  // evaluates to false

	rhs = arrayify(rhs)

	for i, N := 0, rhs.Len(); i < N; i++ {
		if eq(lhs, rhs.Index(i)) {
			return true
		}
	}

	return false
}

func evalBooleanOperator(node *jparse.BooleanOperatorNode, data reflect.Value, env *environment) (reflect.Value, error) {
	// Evaluate both sides and return any errors.
	lhs, err := eval(node.LHS, data, env)
	if err != nil {
		return undefined, err
	}

	// 如果 lhs 为 true，逻辑或的时候，则可以直接返回 true
	if node.Type == jparse.BooleanOr && jlib.Boolean(lhs) {
		return reflect.ValueOf(true), nil
	}

	// 如果 lhs 为 false，逻辑与的时候，则可以直接返回 false
	if node.Type == jparse.BooleanAnd && !jlib.Boolean(lhs) {
		return reflect.ValueOf(false), nil
	}

	rhs, err := eval(node.RHS, data, env)
	if err != nil {
		return undefined, err
	}

	var b bool

	switch node.Type {
	case jparse.BooleanAnd:
		b = jlib.Boolean(lhs) && jlib.Boolean(rhs)
	case jparse.BooleanOr:
		b = jlib.Boolean(lhs) || jlib.Boolean(rhs)
	default:
		panicf("unrecognised boolean operator %q", node.Type)
	}

	return reflect.ValueOf(b), nil
}

func evalStringConcatenation(node *jparse.StringConcatenationNode, data reflect.Value, env *environment) (reflect.Value, error) {
	stringify := func(v reflect.Value) (string, error) {

		if v == undefined || !v.CanInterface() {
			return "", nil
		}
		return jlib.String(v.Interface())
	}

	// Evaluate both sides and return any errors.
	lhs, err := eval(node.LHS, data, env)
	if err != nil {
		return undefined, err
	}

	rhs, err := eval(node.RHS, data, env)
	if err != nil {
		return undefined, err
	}

	// Convert both sides to strings.
	s1, err := stringify(lhs)
	if err != nil {
		return undefined, err
	}

	s2, err := stringify(rhs)
	if err != nil {
		return undefined, err
	}

	return reflect.ValueOf(s1 + s2), nil
}

// Helper functions

func walkObjectValues(v reflect.Value, fn func(reflect.Value)) {
	switch v := jtypes.Resolve(v); {
	case jtypes.IsArray(v):
		for i, N := 0, v.Len(); i < N; i++ {
			fn(v.Index(i))
		}
	case jtypes.IsMap(v):
		for _, k := range v.MapKeys() {
			fn(v.MapIndex(k))
		}
	case jtypes.IsStruct(v):
		for i, N := 0, v.NumField(); i < N; i++ {
			fn(v.Field(i))
		}
	}
}

func normalizeArray(v reflect.Value) reflect.Value {
	v = jtypes.Resolve(v)
	if jtypes.IsArray(v) && v.Len() == 1 {
		return v.Index(0)
	}
	return v
}

func flattenArray(v reflect.Value) reflect.Value {
	results := reflect.MakeSlice(typeInterfaceSlice, 0, 0)

	switch {
	case jtypes.IsArray(v):
		v = jtypes.Resolve(v)
		for i, N := 0, v.Len(); i < N; i++ {
			vi := flattenArray(v.Index(i))
			if vi.IsValid() {
				results = reflect.AppendSlice(results, vi)
			}
		}
	default:
		if v.IsValid() {
			results = reflect.Append(results, v)
		}
	}

	return results
}

func arrayify(v reflect.Value) reflect.Value {
	switch {
	case jtypes.IsArray(v):
		return jtypes.Resolve(v)
	case !v.IsValid():
		return reflect.MakeSlice(typeInterfaceSlice, 0, 0)
	default:
		return reflect.Append(reflect.MakeSlice(typeInterfaceSlice, 0, 1), v)
	}
}

func makeArray(v reflect.Value) reflect.Value {
	switch {
	case jtypes.IsArray(v):
		return jtypes.Resolve(v)
	default:
		arr := reflect.MakeSlice(typeInterfaceSlice, 1, 1)
		if v.IsValid() {
			arr.Index(0).Set(v)
		}
		return arr
	}
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}

// Sequence handling

type Sequence struct {
	Values         []interface{}
	keepSingletons bool
}

func newSequence(size int) *Sequence {
	return &Sequence{
		Values: make([]interface{}, 0, size),
	}
}

func (s *Sequence) Len() int {
	return len(s.Values)
}

func (s *Sequence) Append(v interface{}) {
	s.Values = append(s.Values, v)
}

func (s Sequence) Value() reflect.Value {
	switch n := len(s.Values); {
	case n == 0:
		return undefined
	case n == 1 && !s.keepSingletons:
		return reflect.ValueOf(s.Values[0])
	default:
		return reflect.ValueOf(s.Values)
	}
}

var (
	typeSequence    = reflect.TypeOf((*Sequence)(nil)).Elem()
	typeSequencePtr = reflect.PtrTo(typeSequence)
)

func asSequence(v reflect.Value) (*Sequence, bool) {
	if !v.IsValid() || !v.CanInterface() {
		return nil, false
	}

	if v.Type() == typeSequencePtr {
		return v.Interface().(*Sequence), true
	}

	if jtypes.Resolve(v).Type() == typeSequence && v.CanAddr() {
		return v.Addr().Interface().(*Sequence), true
	}

	return nil, false
}
