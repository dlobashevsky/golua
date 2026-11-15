-- load tests
do
    -- If passed a function load consumes cpu and memory to build the string

    print(runtime.callcontext({kill={memory=10000}}, load, function() return "print('hello')\n" end))
    --> =killed

    print(runtime.callcontext({kill={cpu=10000}}, load, function() return "print('hello')\n" end))
    --> =killed

    -- Same if passed a big string

    print(runtime.callcontext({kill={memory=10000}}, load, ("print('hello')"):rep(1000, "\n")))
    --> =killed

    print(runtime.callcontext({kill={cpu=10000}}, load, ("print('hello')"):rep(10000, "\n")))
    --> =killed
end

-- loadfile tests
do
    -- loadfile consumes cpu and memory to load the string

    local ctx, m = runtime.callcontext({kill={memory=100000}}, loadfile, "lua/big.lua.notest")
    print(ctx, m())
    --> =done	hello

    print(runtime.callcontext({kill={memory=10000}}, loadfile, "lua/big.lua.notest"))
    --> =killed

    local ctx, m = runtime.callcontext({kill={cpu=10000}}, loadfile, "lua/big.lua.notest")
    print(ctx, m())
    --> =done	hello

    print(runtime.callcontext({kill={cpu=1000}}, loadfile, "lua/big.lua.notest"))
    --> =killed
end

-- dofile tests
do
    -- dofile consumes cpu and memory to load the string

    local ctx, m = runtime.callcontext({kill={memory=100000}}, dofile, "lua/big.lua.notest")
    print(ctx, m)
    --> =done	hello

    print(runtime.callcontext({kill={memory=10000}}, dofile, "lua/big.lua.notest"))
    --> =killed

    local ctx, m = runtime.callcontext({kill={cpu=10000}}, dofile, "lua/big.lua.notest")
    print(ctx, m)
    --> =done	hello

    print(runtime.callcontext({kill={cpu=1000}}, dofile, "lua/big.lua.notest"))
    --> =killed
end

-- With only a time limit, CPU is counted but there is no CPU budget: loadfile
-- and dofile must read the whole file (they used to read nothing once some CPU
-- had been used, as the unused CPU wrapped around).  A soft time limit tracks
-- CPU the same way as a hard one, without requiring "timesafe" from loadfile.
do
    local function spin() local n = 0 for i = 1, 1000 do n = n + i end return n end

    print(runtime.callcontext({stop={millis=10000}}, function() spin() return dofile("lua/big.lua.notest") end))
    --> =done	hello

    local ctx, m = runtime.callcontext({stop={millis=10000}}, function() spin() return loadfile("lua/big.lua.notest") end)
    print(ctx, m())
    --> =done	hello

    print(runtime.callcontext({stop={millis=10000}, kill={memory=1000000}}, function() spin() return dofile("lua/big.lua.notest") end))
    --> =done	hello
end

-- A long pattern match under a time limit alone still reports its work, so the
-- time limit stops it (the budget without a CPU limit is large, not "none").
do
    local s = ("a"):rep(50000000)
    print(runtime.callcontext({kill={millis=1}}, string.find, s, "a*$"))
    --> =killed
end
