package lightweigit

//go:generate bash -c "rm -rf target/* tmp/*"

//go:generate go run github.com/amazing-generators/gometagen/cmd/gometagen@latest generate -source _run/values.yml -hash-source . -hash-exclude .git -hash-exclude .idea -hash-exclude target -hash-exclude tmp -format go -out target/meta_gen.go -pkg target -force
//go:generate go run ./_generate/build_map
//go:generate go run ./_generate/build_func
