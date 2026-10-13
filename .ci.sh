#! /bin/sh

set -eu

run() {
  echo "  -> ${@}"

  if output=$("${@}" 2>&1); then
    return 0
  else
    status=${?}
    echo "${output}"
    return "${status}"
  fi
}

echo '=> Templ'
(
  cd backend
  go tool templ generate ./internal/frontend/
)

for dir in backend smtp_relay; do
  (
    echo "=> ${dir}"
    cd "${dir}"

    run go vet "${@}" ./...
    run go fix -diff "${@}" ./...
    run go test "${@}" ./...

    if test -n "${CI:-}"; then
      run gofmt -d .
    else
      go mod tidy
      run go fmt ./...
    fi
  )
done
