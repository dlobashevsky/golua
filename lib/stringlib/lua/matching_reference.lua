-- Results checked against the reference implementation (PUC-Rio Lua 5.4.6).
-- gsub: an anchored pattern matches at most once, at the start (arnodel/golua#239)
print(string.gsub("abc", "^", "x"))
--> =xabc	1
print(string.gsub("aaa", "^a", "b"))
--> =baa	1
print(string.gsub(" x ", "^%s+", ""))
--> =x 	1
print(string.gsub("xab", "^a", "-"))
--> =xab	0
print(string.gsub("aaa", "^a*$", "-"))
--> =-	1
print(string.gsub("aaa", "^a", "-", 5))
--> =-aa	1
print(string.gsub("hello", "^(h)", string.upper))
--> =Hello	1
print(string.gsub("abc", "^(a)", {a = "A"}))
--> =Abc	1
-- gsub: n <= 0 makes no substitution; nil n is no limit
print(string.gsub("aaa", "a", "b", 0))
--> =aaa	0
print(string.gsub("aaa", "a", "b", -1))
--> =aaa	0
print(string.gsub("aaa", "a", "b", nil))
--> =bbb	3
-- gsub: an empty match right after a match does not count
print(string.gsub("abc", "%w*", "-"))
--> =-	1
print(string.gsub("hello world", "%a*", "<%0>"))
--> =<hello> <world>	2
-- find: plain search and the empty pattern give positions in s (arnodel/golua#240)
print(string.find("abcabc", "b", 3, true))
--> =5	5
print(string.find("abcabc", "", 3))
--> =3	2
print(string.find("abc", "", 4))
--> =4	3
print(string.find("abc", "", 5))
--> =nil
print(string.find("a.b.c", ".", 3, true))
--> =4	4
print(string.find("abc", "b", nil, true))
--> =2	2
-- match: an init past the end is no match, not a crash
print(string.match("", "^", 2))
--> =nil
print(string.match("abc", "()", 5))
--> =nil
print(string.match("abc", "()", 4))
--> =4
-- gmatch: '^' is an ordinary character, an exhausted iterator returns nothing (arnodel/golua#241)
do
    local out = {}
    for w in string.gmatch("a^b", "^") do out[#out + 1] = w end
    print(#out, out[1])
    --> =1	^
    out = {}
    for w in string.gmatch("^b a^b", "^b") do out[#out + 1] = w end
    print(#out, out[1], out[2])
    --> =2	^b	^b
    print(select("#", string.gmatch("abc", "x")()))
    --> =0
    print(select("#", string.gmatch("abc", "", 5)()))
    --> =0
    local it = string.gmatch("abc", "", 4)
    print(select("#", it()), select("#", it()))
    --> =1	0
    out = {}
    for w in string.gmatch("abc", "%w*") do out[#out + 1] = "<" .. w .. ">" end
    print(table.concat(out))
    --> =<abc>
end
-- rep: a huge count is a Lua error, not a Go panic (arnodel/golua#242)
print(pcall(string.rep, "", 1 << 62, "x"))
--> ~^false	.*resulting string too large$
print(pcall(string.rep, "x", 1 << 62))
--> ~^false	.*resulting string too large$
print(string.rep("ab", 3, ","), string.rep("ab", 0, ","), string.rep("ab", -5))
--> =ab,ab,ab		
-- gsub: substitutions that produce nothing still change the string
print(string.gsub("aaa", "a", ""))
--> =	3
print(string.gsub("   ", "%s", ""))
--> =	3
print(string.gsub("abc", "%w", {}))
--> =abc	3
