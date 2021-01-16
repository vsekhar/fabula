package youtime

import (
	"math"
	"testing"

	"github.com/pa-m/sklearn/svm"
	"gonum.org/v1/gonum/mat"
)

type counterMatrix struct {
	m     mat.Matrix
	reads int
	ts    int
}

func (m *counterMatrix) Dims() (r, c int) {
	return m.m.Dims()
}

func (m *counterMatrix) At(i, j int) float64 {
	m.reads++
	return m.m.At(i, j)
}

func (m *counterMatrix) T() mat.Matrix {
	m.ts++
	return m.m.T()
}

func (m *counterMatrix) RowView(i int) mat.Vector {
	return m.m.(mat.RowViewer).RowView(i)
}

func TestSvm(t *testing.T) {
	// from: https://stats.stackexchange.com/questions/39243/how-does-one-interpret-svm-feature-weights
	x := &counterMatrix{
		m: mat.NewDense(2, 2, []float64{
			// x1, x2
			0, 0,
			-4, 4,
		}),
	}
	y := &counterMatrix{
		m: mat.NewDense(2, 1, []float64{
			1,
			-1,
		}),
	}

	t.Logf("X: %+v", x)
	t.Logf("Y: %+v", y)

	// TODO: input normalization?

	clf := svm.NewSVC()
	clf.Kernel = "linear"
	clf.MaxIter = 20

	// C is cost of misclassified examples. Lower C gives wider margin and more
	// examples in the roadway. Higher C produces narrower margin and fewer
	// examples in the roadway. Typical values are 0.1 .. 100.
	//
	// We remove impure probes from our data set, so this shuold skew high.
	//
	// See: https://stats.stackexchange.com/questions/31066/what-is-the-influence-of-c-in-svms-with-linear-kernel.
	clf.C = 1

	clf.Fit(x, y)
	if len(clf.Model) != 1 {
		t.Fatalf("Expected 1 model, got %d", len(clf.Model))
	}
	model := clf.Model[0]
	t.Logf("Model[0]: %+v", model)
	x2 := mat.NewDense(2, 2, []float64{
		1, 1,
		47, 48,
	})
	y2 := clf.Predict(x2, nil)
	t.Logf("Y2: %+v", y2)
	t.Logf("Score: %f", clf.Score(x2, y2))

	if len(model.Alphas) != len(model.Support) {
		t.Fatalf("did not get the same number of alphas (%d) as support vectors (%d)", len(model.Alphas), len(model.Support))
	}
	w := boundaryNormal(x, y, model)
	t.Logf("W: %+v", w)
	m, b, above := decision(w, model.B)
	if above {
		t.Logf("H(x): x_k >= %v*x_ + %.2f", mat.Formatted(m), b)
	} else {
		t.Logf("H(x): x_k <= %v*x_ + %.2f", mat.Formatted(m), b)
	}
	em, eb, eabove := mat.NewVecDense(1, []float64{1.00}), 4., false
	if !mat.Equal(em, m) || eb != b || eabove != above {
		t.Errorf("Expected decision function %v*x_ + %.2f, got %v*x_ + %.2f", mat.Formatted(em), eb, mat.Formatted(m), b)
	}
	r := radius(w)
	er := 4 / math.Sqrt(2.0)
	if r != er {
		t.Errorf("Expected radius %.2f, got %.2f", er, r)
	}
	t.Logf("Reads: x=%d, y=%d", x.reads, y.reads)
	t.Logf("Transposes: x=%d, y=%d", x.ts, y.ts)
	t.Error("Output above")
}
