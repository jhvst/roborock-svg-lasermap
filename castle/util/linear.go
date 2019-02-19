package util

import (
	"bufio"
	"bytes"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/sajari/regression"
)

func ConvertToBoundingBox(x, y float64) BeamBoundingBox {
	hypotenuse := math.Sqrt(
		math.Pow(x, 2) + math.Pow(y, 2),
	)
	angle := math.Acos(x/hypotenuse) * 180 / math.Pi
	bbb, x, y := BoundingBox(hypotenuse, angle)
	return bbb
}

func BoundingBox(hypotenuse, angle float64) (thisBeam BeamBoundingBox, x, y float64) {
	angle_rad := angle*(2*math.Pi)/360;

	// Scale box width and height (depth). Resolution = 1 deg.
	bounding_side_d := 1.0 * math.Sin(1*math.Pi/180) * hypotenuse; //(1.23)
	bounding_side_w := 1.0 * bounding_side_d; //(1.45)

        // Box centre
	x = math.Cos(angle_rad)*hypotenuse;
	y = math.Sin(angle_rad)*hypotenuse;

	// Dims
	bwx := math.Cos(math.Pi/2+angle_rad)*bounding_side_w;
	bwy := math.Sin(math.Pi/2+angle_rad)*bounding_side_w;
	blx := math.Cos(angle_rad)*bounding_side_d;
	bly := math.Sin(angle_rad)*bounding_side_d;

	// Corner stones
	thisBeam = BeamBoundingBox{
		UpperRight: BeamPoint{}.New(x+bwx+blx, y+bwy+bly),
		LowerRight: BeamPoint{}.New(x+bwx-blx, y+bwy-bly),
		UpperLeft:  BeamPoint{}.New(x-bwx+blx, y-bwy+bly),
		LowerLeft:  BeamPoint{}.New(x-bwx-blx, y-bwy-bly),
	}

	return thisBeam, x, y
}

// Regression takes in a CSV data, which should have
// "link" suffixes added. The regression will iterate over
// the given data and replace linked parts with linear models,
// which should be connected with a point.
func Regression(data []byte) (map[int][]float64, error) {
	var connectedBuffer bytes.Buffer
	cScanner := bufio.NewScanner(bytes.NewReader(
		bytes.TrimSpace(data),
	))
	lc := 0
	angleMap := make(map[int][]float64)
	for cScanner.Scan() {

		if lc == 360 {
			lc = 0
		}

		if strings.HasSuffix(cScanner.Text(), "link") {
			connectedBuffer.WriteString(cScanner.Text() + "\n")
			lc++
			continue
		}

		if connectedBuffer.Len() > 0 {

			r := new(regression.Regression)
			r.SetVar(0, "x")

			eScanner := bufio.NewScanner(bytes.NewReader(
				bytes.TrimSpace(connectedBuffer.Bytes()),
			))

			var ymin float64 = math.MaxFloat64
			var ymax float64 = math.MinInt64
			for eScanner.Scan() {

				values := strings.Split(eScanner.Text(), ",")

				var boundingbox string
				for _, s := range values[3] {
					switch s {
					case '}', '{':
						continue
					default:
						boundingbox = boundingbox + string(s)
					}
				}

				var edges [][]float64
				var point []float64
				for _, edge := range strings.Split(boundingbox, " ") {
					e, _ := strconv.ParseFloat(edge, 64)
					point = append(point, e)
					if len(point) == 2 {
						edges = append(edges, point)
						point = []float64{}
					}
				}

				y, _ := strconv.ParseFloat(values[5], 64)
				if y < ymin {
					ymin = y
				}
				if y > ymax {
					ymax = y
				}

				r.Train(regression.MakeDataPoints(edges, 0)...)
			}
			if err := eScanner.Err(); err != nil {
				return angleMap, err
			}

			err := r.Run()
			if err != nil {
				return angleMap, err
			}

			var formulaBuffer bytes.Buffer
			for _, r := range strings.TrimLeft(r.Formula, "Predicted = ") {
				switch r {
				case ' ', '+', 'x':
					continue
				case '*':
					formulaBuffer.WriteString(",")
				default:
					formulaBuffer.WriteString(string(r))
				}
			}
			formula := strings.Split(formulaBuffer.String(), ",")
			a, _ := strconv.ParseFloat(formula[0], 64)
			b, _ := strconv.ParseFloat(formula[1], 64)

			lineStart := a + ymin*b
			lineEnd := a + ymax*b

			if val, ok := angleMap[lc]; ok {

				s := []float64{
					(val[0] / lineStart),
					(val[1] / ymin),
					(val[2] / lineEnd),
					(val[3] / ymax),
				}
				sort.Float64s(s)

				if s[0] > 0.95 && s[3] < 1.05 {
					lineStart = math.Min(val[0], lineStart)
					ymin = math.Min(val[1], ymin)
					lineEnd = math.Max(val[2], lineEnd)
					ymax = math.Max(val[3], ymax)
				}
			}

			angleMap[lc] = []float64{lineStart, ymin, lineEnd, ymax}
		}

		lc++
		connectedBuffer.Reset()
	}

	return angleMap, cScanner.Err()
}
