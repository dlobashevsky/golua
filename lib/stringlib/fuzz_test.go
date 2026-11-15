package stringlib_test

import (
	"strings"
	"testing"
	"time"

	"github.com/arnodel/golua/lib"
	rt "github.com/arnodel/golua/runtime"
)

// FuzzPatternFunctions runs string.find, match, gmatch, gsub and rep on arbitrary subjects,
// patterns and inits. Lua errors (invalid patterns, ...) are fine; a Go panic is not: it would
// escape pcall and take the embedding program down.
func FuzzPatternFunctions(f *testing.F) {
	for _, seed := range []struct {
		s, p string
		i    int64
	}{
		{"hello world", "%w+", 1},
		{"", "^", 2},
		{"abc", "^(a)(b)()", 0},
		{"a^b", "^", 5},
		{"xx", "%bxy", -1},
		{"  trim  ", "^%s*(.-)%s*$", 1},
		{"a.b", "%f[%w]%w+", 1},
		{"aaa", "a-$", 100},
	} {
		f.Add(seed.s, seed.p, seed.i)
	}
	r := rt.New(nil)
	defer r.Close(nil)
	cleanup := lib.LoadAll(r)
	defer cleanup()
	chunk, err := r.CompileAndLoadLuaChunk("fuzz", []byte(`
local s, p, i = ...
pcall(string.find, s, p, i)
pcall(string.find, s, p, i, true)
pcall(string.match, s, p, i)
pcall(function()
    local n = 0
    for _ in string.gmatch(s, p, i) do
        n = n + 1
        if n > 1000 then break end
    end
end)
pcall(string.gsub, s, p, "%0%0")
pcall(string.gsub, s, p, "", i)
pcall(string.gsub, s, p, function(...) return select("#", ...) end)
pcall(string.gsub, s, p, {})
pcall(string.rep, s, i, p)
`), rt.TableValue(r.GlobalEnv()))
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, s, p string, i int64) {
		if len(s) > 1000 || len(p) > 100 {
			return
		}
		if i > 1<<20 || i < -(1<<20) {
			// Keep string.rep small: a huge count is an error or a big allocation, not a bug.
			i %= 1 << 20
		}
		var def rt.RuntimeContextDef
		def.HardLimits.Cpu = 10_000_000
		def.HardLimits.Memory = 100_000_000
		th := r.MainThread()
		start := time.Now()
		defer func() {
			if d := time.Since(start); d > 2*time.Second {
				t.Fatalf("s=%q p=%q i=%d: took %s under a CPU limit", s, p, i, d)
			}
		}()
		_, err := th.CallContext(def, func() error {
			return rt.Call(th, rt.FunctionValue(chunk), []rt.Value{rt.StringValue(s), rt.StringValue(p), rt.IntValue(i)}, rt.NewTerminationWith(nil, 0, false))
		})
		if err != nil && !strings.Contains(err.Error(), "limit") {
			t.Fatalf("s=%q p=%q i=%d: %v", s, p, i, err)
		}
	})
}
