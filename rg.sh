#!/bin/bash
# Ripgrep-like search tool (using grep for now)
# Usage: ./rg.sh <pattern> [path]

if [ $# -eq 0 ]; then
    echo "Usage: $0 <pattern> [path]"
    exit 1
fi

pattern="$1"
path="${2:-.}"

# Recursive grep with some nice options
grep -r -n -I --color=always "$pattern" "$path" 2>/dev/null || echo "No matches found"
