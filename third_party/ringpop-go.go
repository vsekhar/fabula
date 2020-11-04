package thirdparty

//go:generate rm -rf github.com/uber/ringpop-go
//go:generate git clone --depth=1 https://github.com/uber/ringpop-go.git github.com/uber/ringpop-go/
//go:generate sh -c "git -C github.com/uber/ringpop-go rev-parse HEAD > github.com/uber/ringpop-go/git.sum"
//go:generate rm -rf github.com/uber/ringpop-go/.git
