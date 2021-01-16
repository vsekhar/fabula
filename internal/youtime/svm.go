package youtime

import (
	"math"

	"github.com/pa-m/sklearn/svm"
	"gonum.org/v1/gonum/mat"
)

// TODO: simple 2D linear support vector machine

// Resources:
// * gonum.org/v1/gonum/optimize
// * github.com/pa-m/sklearn/svm

// boundaryNormal returns the boundary normal vector for model fitted on x and
// y. This vector is often denoted w.
//
// boundaryNormal runs in O(d) complexity, where d is the number of dimensions
// (features) in the data set x.
func boundaryNormal(x, y mat.Matrix, model *svm.Model) mat.Vector {
	// w = sum(alpha_i * y_i * x_i) // for support vectors only
	_, c := x.Dims()
	w := mat.NewVecDense(c, nil)
	for i, sv := range model.Support {
		xi := x.(mat.RowViewer).RowView(sv)
		yi := y.At(sv, 0)
		alpha := model.Alphas[i]
		w.AddScaledVec(w, alpha*yi, xi)
	}
	return w
}

// radius returns the radius around the decision boundary specified by the
// boundary normal vector w.
func radius(w mat.Vector) float64 {
	return 1.0 / math.Sqrt(mat.Dot(w, w))
}

// decision returns the coefficients of the decision function for a given
// boundary normal vector w and intercept bm with the last dimension k expressed
// as a function of all dimensions 1 through k-1. With x_ as x without dimension
// k:
//
//  Boundary: x_k = m*x_ + b
//
// If above is true, the decision function is true if x_k is above the boundary.
// Otherwise the decision function is true if x_k is below the boundary.
//
//  Decision:
//    H(x): x_k >= m*x_ + b for above==true
//    H(x): x_k <= m*x_ + b for above==false
func decision(w mat.Vector, bm float64) (m mat.Vector, b float64, above bool) {
	//  x_k = (-w_1 / w_k)*x_1 + (-w_2 / w_k)*x_2 + ... + (-b_m / w_k)
	//      = (-w_x_ - b_m) / w_k
	//
	// Where w_ and x_ are w and x without dimension k.
	//
	//  m = -w_ / w_k = w_* / (-w_k)
	//  b = -b_m / w_k = b_m / (-w_k)

	wK := w.At(w.Len()-1, 0)
	wRest := w.(*mat.VecDense).SliceVec(0, w.Len()-1) // w_
	dm := mat.NewVecDense(w.Len()-1, nil)
	dm.ScaleVec(1/(-wK), wRest)
	m = dm
	b = bm / (-wK)
	above = wK > 0
	return
}

// TODO: lane width
