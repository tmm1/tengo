package stdlib_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/d5/tengo/assert"
	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/stdlib"
)

type ARR = []interface{}
type MAP = map[string]interface{}
type IARR []interface{}
type IMAP map[string]interface{}

func TestAllModuleNames(t *testing.T) {
	names := stdlib.AllModuleNames()
	if !assert.Equal(t, len(stdlib.Modules), len(names)) {
		return
	}
	for _, name := range names {
		assert.NotNil(t, stdlib.Modules[name], "name: %s", name)
	}
}

func TestGetModules(t *testing.T) {
	mods := stdlib.GetModules()
	assert.Equal(t, 0, len(mods))

	mods = stdlib.GetModules("os")
	assert.Equal(t, 1, len(mods))
	assert.NotNil(t, mods["os"])

	mods = stdlib.GetModules("os", "rand")
	assert.Equal(t, 2, len(mods))
	assert.NotNil(t, mods["os"])
	assert.NotNil(t, mods["rand"])

	mods = stdlib.GetModules("text", "text")
	assert.Equal(t, 1, len(mods))
	assert.NotNil(t, mods["text"])

	mods = stdlib.GetModules("nonexisting", "text")
	assert.Equal(t, 1, len(mods))
	assert.NotNil(t, mods["text"])
}

type callres struct {
	t *testing.T
	o objects.Object
	e error
}

func (c callres) call(funcName string, args ...interface{}) callres {
	if c.e != nil {
		return c
	}

	imap, ok := c.o.(*objects.ImmutableMap)
	if !ok {
		return c
	}

	m, ok := imap.Value[funcName]
	if !ok {
		return callres{t: c.t, e: fmt.Errorf("function not found: %s", funcName)}
	}

	f, ok := m.(*objects.UserFunction)
	if !ok {
		return callres{t: c.t, e: fmt.Errorf("non-callable: %s", funcName)}
	}

	var oargs []objects.Object
	for _, v := range args {
		oargs = append(oargs, object(v))
	}

	res, err := f.Value(oargs...)

	return callres{t: c.t, o: res, e: err}
}

func (c callres) expect(expected interface{}, msgAndArgs ...interface{}) bool {
	return assert.NoError(c.t, c.e, msgAndArgs...) &&
		assert.Equal(c.t, object(expected), c.o, msgAndArgs...)
}

func (c callres) expectError() bool {
	return assert.Error(c.t, c.e)
}

func module(t *testing.T, moduleName string) callres {
	mod, ok := stdlib.Modules[moduleName]
	if !ok {
		return callres{t: t, e: fmt.Errorf("module not found: %s", moduleName)}
	}

	return callres{t: t, o: (*mod).(*objects.ImmutableMap)}
}

func object(v interface{}) objects.Object {
	switch v := v.(type) {
	case objects.Object:
		return v
	case string:
		return &objects.String{Value: v}
	case int64:
		return &objects.Int{Value: v}
	case int: // for convenience
		return &objects.Int{Value: int64(v)}
	case bool:
		if v {
			return objects.TrueValue
		}
		return objects.FalseValue
	case rune:
		return &objects.Char{Value: v}
	case byte: // for convenience
		return &objects.Char{Value: rune(v)}
	case float64:
		return &objects.Float{Value: v}
	case []byte:
		return &objects.Bytes{Value: v}
	case MAP:
		objs := make(map[string]objects.Object)
		for k, v := range v {
			objs[k] = object(v)
		}

		return &objects.Map{Value: objs}
	case ARR:
		var objs []objects.Object
		for _, e := range v {
			objs = append(objs, object(e))
		}

		return &objects.Array{Value: objs}
	case IMAP:
		objs := make(map[string]objects.Object)
		for k, v := range v {
			objs[k] = object(v)
		}

		return &objects.ImmutableMap{Value: objs}
	case IARR:
		var objs []objects.Object
		for _, e := range v {
			objs = append(objs, object(e))
		}

		return &objects.ImmutableArray{Value: objs}
	case time.Time:
		return &objects.Time{Value: v}
	case []int:
		var objs []objects.Object
		for _, e := range v {
			objs = append(objs, &objects.Int{Value: int64(e)})
		}

		return &objects.Array{Value: objs}
	}

	panic(fmt.Errorf("unknown type: %T", v))
}