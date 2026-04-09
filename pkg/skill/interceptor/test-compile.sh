#!/bin/bash

# Script to check if interceptor implementations compile correctly

echo "Checking interceptor implementations..."

# Check auth interceptor
echo -n "Auth interceptor: "
if go build -o /dev/null ./auth/ 2>/dev/null; then
    echo "✅ Compiled successfully"
else
    echo "❌ Failed to compile"
fi

# Check paramcheck interceptor
echo -n "Paramcheck interceptor: "
if go build -o /dev/null ./paramcheck/ 2>/dev/null; then
    echo "✅ Compiled successfully"
else
    echo "❌ Failed to compile"
fi

# Check if tests pass
echo -n "Running tests: "
if go test ./auth/ ./paramcheck/ -v; then
    echo "✅ All tests passed"
else
    echo "❌ Some tests failed"
fi

echo "Done."