package graphql2

//go:generate rm -f mapconfig.go
//go:generate rm -f maplimit.go
//go:generate go run github.com/99designs/gqlgen -config gqlgen.yml
//go:generate gofmt -s -w generated.go
//go:generate gofmt -s -w models_gen.go
//go:generate go run ../devtools/configparams -out mapconfig.go
//go:generate go run ../devtools/limitapigen -out maplimit.go
//go:generate go run golang.org/x/tools/cmd/goimports -w mapconfig.go
//go:generate go run golang.org/x/tools/cmd/goimports -w maplimit.go
