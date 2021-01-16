package youtime

import (
	"testing"

	"gonum.org/v1/gonum/mat"
)

// TODO: test ProbeStore against libsvm (see svm_test.go) and liblinear.

/*
func train(x *featureMatrix, y []float64) *C.struct_model {
	problem := &C.struct_problem{
		// x: x.asCArray(),
		// y:    (*C.double)(&y[0]),
		l:    C.int(len(x.nodes)),
		n:    C.int(x.c),
		bias: 1.0,
	}
	params := &C.struct_parameter{
		solver_type: C.L2R_L2LOSS_SVC,
		eps:         0.1,
		C:           1.0,
		nr_weight:   0,   // symmetric/balanced input data
		p:           0.5, // or 2?
	}
	cchk := C.check_parameter(problem, params)
	chk := C.GoString(cchk)
	if chk != "" {
		panic(chk)
	}
	return C.train(problem, params)
}
*/

// Should we just copy probe data into C arrays? Yes. We can move 1000 probe
// values in 4 microseconds.
func BenchmarkCopyingMatrix(b *testing.B) {
	m := mat.NewDense(b.N, 2, nil)
	// prevent optimizer
	for i := 0; i < b.N; i++ {
		m.SetRow(i, []float64{float64(i), float64(i + 1)})
	}
	// This is comparable to copying into a C array
	// m2 := mat.DenseCopyOf(m)
	m2 := mat.NewDense(b.N, 2, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m2.Set(i, 0, m.At(i, 0))
		m2.Set(i, 1, m.At(i, 1))
	}
	r2, c2 := m2.Dims()
	if r, c := m.Dims(); r2 != r || c2 != c {
		b.Error("copy failed")
	}
}
