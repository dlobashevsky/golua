package stringlib

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/arnodel/golua/lib/packagelib"
	"github.com/arnodel/golua/luastrings"
	rt "github.com/arnodel/golua/runtime"
)

// LibLoader specifies how to load the string lib
var LibLoader = packagelib.Loader{
	Load: load,
	Name: "string",
}

func load(r *rt.Runtime) (rt.Value, func()) {
	pkg := rt.NewTable()
	pkgVal := rt.TableValue(pkg)

	rt.SolemnlyDeclareCompliance(
		rt.ComplyCpuSafe|rt.ComplyMemSafe|rt.ComplyTimeSafe|rt.ComplyIoSafe,

		r.SetEnvGoFunc(pkg, "byte", bytef, 3, false),
		r.SetEnvGoFunc(pkg, "char", char, 0, true),
		r.SetEnvGoFunc(pkg, "dump", dump, 2, false),
		r.SetEnvGoFunc(pkg, "find", find, 4, false),
		r.SetEnvGoFunc(pkg, "gmatch", gmatch, 3, false),
		r.SetEnvGoFunc(pkg, "gsub", gsub, 4, false),
		r.SetEnvGoFunc(pkg, "len", lenf, 1, false),
		r.SetEnvGoFunc(pkg, "lower", lower, 1, false),
		r.SetEnvGoFunc(pkg, "match", match, 3, false),
		r.SetEnvGoFunc(pkg, "upper", upper, 1, false),
		r.SetEnvGoFunc(pkg, "rep", rep, 3, false),
		r.SetEnvGoFunc(pkg, "reverse", reverse, 1, false),
		r.SetEnvGoFunc(pkg, "sub", sub, 3, false),
		r.SetEnvGoFunc(pkg, "format", format, 1, true),
		r.SetEnvGoFunc(pkg, "pack", pack, 1, true),
		r.SetEnvGoFunc(pkg, "packsize", packsize, 1, false),
		r.SetEnvGoFunc(pkg, "unpack", unpack, 3, false),
	)

	stringMeta := rt.NewTable()
	r.SetEnv(stringMeta, "__index", pkgVal)

	rt.SolemnlyDeclareCompliance(
		rt.ComplyCpuSafe|rt.ComplyMemSafe|rt.ComplyTimeSafe|rt.ComplyIoSafe,

		r.SetEnvGoFunc(stringMeta, "__add", string__add, 2, false),
		r.SetEnvGoFunc(stringMeta, "__sub", string__sub, 2, false),
		r.SetEnvGoFunc(stringMeta, "__mul", string__mul, 2, false),
		r.SetEnvGoFunc(stringMeta, "__div", string__div, 2, false),
		r.SetEnvGoFunc(stringMeta, "__idiv", string__idiv, 2, false),
		r.SetEnvGoFunc(stringMeta, "__mod", string__mod, 2, false),
		r.SetEnvGoFunc(stringMeta, "__pow", string__pow, 2, false),
		r.SetEnvGoFunc(stringMeta, "__unm", string__unm, 1, false),
	)
	r.SetStringMeta(stringMeta)

	return pkgVal, nil
}

func maxpos(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func minpos(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func bytef(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	s, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	i, j := 1, 1
	if c.NArgs() >= 2 {
		ii, err := c.IntArg(1)
		if err != nil {
			return nil, err
		}
		i = luastrings.StringNormPos(s, rt.ClampToInt(ii))
		j = i
	}
	if c.NArgs() >= 3 {
		jj, err := c.IntArg(2)
		if err != nil {
			return nil, err
		}
		j = luastrings.StringNormPos(s, rt.ClampToInt(jj))
	}
	next := c.Next()
	i = maxpos(1, i)
	j = minpos(len(s), j)
	for i <= j {
		t.Push1(next, rt.IntValue(int64(s[i-1])))
		i++
	}
	return next, nil
}

func char(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	vals := c.Etc()
	buf := make([]byte, len(vals))
	for i, v := range vals {
		x, ok := rt.ToInt(v)
		if !ok {
			return nil, errors.New("arguments must be integers")
		}
		if x < 0 || x > 255 {
			return nil, fmt.Errorf("#%d out of range", i+1)
		}
		buf[i] = byte(x)
	}
	t.RequireBytes(len(buf))
	return c.PushingNext1(t.Runtime, rt.StringValue(string(buf))), nil
}

func lenf(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	s, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	return c.PushingNext1(t.Runtime, rt.IntValue(int64(len(s)))), nil
}

func lower(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	s, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	t.RequireBytes(len(s))
	s = strings.ToLower(string(s))
	return c.PushingNext1(t.Runtime, rt.StringValue(s)), nil
}

func upper(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	s, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	t.RequireBytes(len(s))
	s = strings.ToUpper(string(s))
	return c.PushingNext1(t.Runtime, rt.StringValue(s)), nil
}

func rep(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	ls, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	ln, err := c.IntArg(1)
	if err != nil {
		return nil, err
	}
	var sep string
	if c.NArgs() >= 3 && !c.Arg(2).IsNil() {
		sep, err = c.StringArg(2)
		if err != nil {
			return nil, err
		}
	}
	// As str_rep in the reference implementation: a count <= 0 gives the empty
	// string, and a result longer than maxRepSize is an error.  The check comes
	// before any allocation, so a huge count is a Lua error rather than a Go
	// panic in makeslice.
	if ln <= 0 {
		return c.PushingNext1(t.Runtime, rt.StringValue("")), nil
	}
	unit := int64(len(ls)) + int64(len(sep))
	if unit > maxRepSize/ln {
		return nil, errors.New("resulting string too large")
	}
	n := int(ln)
	if n == 1 {
		return c.PushingNext1(t.Runtime, rt.StringValue(ls)), nil
	}
	sz := n*len(ls) + (n-1)*len(sep)
	t.RequireBytes(sz)
	if sep == "" {
		return c.PushingNext1(t.Runtime, rt.StringValue(strings.Repeat(ls, n))), nil
	}
	builder := strings.Builder{}
	builder.Grow(sz)
	builder.WriteString(ls)
	for i := 1; i < n; i++ {
		builder.WriteString(sep)
		builder.WriteString(ls)
	}
	return c.PushingNext1(t.Runtime, rt.StringValue(builder.String())), nil
}

// maxRepSize is the longest string string.rep makes: MAXSIZE (INT_MAX) in the
// reference implementation.
const maxRepSize = math.MaxInt32

func reverse(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	s, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	t.RequireBytes(len(s))
	sb := []byte(s)
	l := len(s) - 1
	for i := 0; 2*i <= l; i++ {
		sb[i], sb[l-i] = sb[l-i], sb[i]
	}
	return c.PushingNext1(t.Runtime, rt.StringValue(string(sb))), nil
}

func sub(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	s, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	ii, err := c.IntArg(1)
	if err != nil {
		return nil, err
	}
	i := luastrings.StringNormPos(s, rt.ClampToInt(ii))
	j := len(s)
	if c.NArgs() >= 3 {
		jj, err := c.IntArg(2)
		if err != nil {
			return nil, err
		}
		j = luastrings.StringNormPos(s, rt.ClampToInt(jj))
	}
	var slice string
	i = maxpos(1, i)
	j = minpos(len(s), j)
	if i <= len(s) && i <= j {
		t.RequireBytes(j - i + 1)
		slice = s[i-1 : j]
	}
	return c.PushingNext1(t.Runtime, rt.StringValue(slice)), nil
}
