# Lua Mirrorr Server

proxy and cache needed package from luarock

```bash
build:

$ make

usage:

$ ./luamirror --help
usage: luamirror [<flags>]

Flags:
-h, --help               Show context-sensitive help (also try --help-long and --help-man).
    --web.listen-address=":8080"  
                         The address to listen on for web interface.
    --luarock.server="http://luafr.org/luarocks"  
                         Luarock server address
    --server.timeout=5s  Timeout for luarock server
    --data.dir="."       Data directory
    --version            Show application version.

```