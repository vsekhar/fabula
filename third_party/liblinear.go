package thirdparty

//go:generate rm -rf github.com/cjlin1/liblinear
//go:generate git clone --depth=1 https://github.com/cjlin1/liblinear.git github.com/cjlin1/liblinear/
//go:generate sh -c "git -C github.com/cjlin1/liblinear rev-parse HEAD > github.com/cjlin1/liblinear/git.sum"
//go:generate rm -rf github.com/cjlin1/liblinear/.git
//go:generate make -j8 -C github.com/cjlin1/liblinear lib
