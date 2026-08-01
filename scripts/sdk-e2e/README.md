# SDK end-to-end smoke tests

Against a live Rexec API (default: https://rexec.sh guest session).

```bash
# Obtain a guest token
export URL=https://rexec.sh
export TOKEN=$(curl -sS -X POST "$URL/api/auth/guest" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"e2e$RANDOM\",\"email\":\"e2e$RANDOM@test.com\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

# JS (local build)
(cd ../../sdk/js && npm run build)
node test-js.mjs   # set URL/TOKEN in env or .env

# Python
URL=$URL TOKEN=$TOKEN python3 test_py.py

# Go
(cd go && go run .)

# Rust
URL=$URL TOKEN=$TOKEN cargo run --manifest-path rust_e2e/Cargo.toml

# Ruby (Ruby >= 3.0 + faraday)
URL=$URL TOKEN=$TOKEN ruby test_rb.rb

# .NET
URL=$URL TOKEN=$TOKEN dotnet run --project dotnet_e2e -f net9.0
```

Flow exercised: **list → create(image: ubuntu) → list → get → delete**.
