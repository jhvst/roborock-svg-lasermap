package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"../castle/util"
	"github.com/sajari/regression"
)

type Point struct {
	X, Y       float64
	Hypotenuse float64
	Angle      float64
}

func (p Point) GetSide() Side {
	// if X is positive, it is either TopRight or LowerLeft
	if p.X > 0 {
		if p.Y > 0 {
			return TopRight
		}
		return LowerLeft
	}
	// if X is negative, it is either TopLeft or LowerLeft
	if p.Y > 0 {
		return TopLeft
	}
	return LowerLeft
}

type Side uint

const (
	TopRight   Side = 0
	TopLeft    Side = 1
	LowerLeft  Side = 2
	LowerRight Side = 3
)

type Floorplan struct {
	pointsRaw []Point

	bounds map[Side][]Point
}

func (f *Floorplan) Load(filename string) error {
	f.bounds = make(map[Side][]Point)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, point := range records {
		x, err := strconv.ParseFloat(point[0], 64)
		if err != nil {
			return err
		}
		y, err := strconv.ParseFloat(point[1], 64)
		if err != nil {
			return err
		}

		p := Point{X: x, Y: y}
		f.pointsRaw = append(f.pointsRaw, p)

		side := p.GetSide()
		f.bounds[side] = append(f.bounds[side], p)
	}

	return nil
}

func (p Point) dist(target Point) float64 {
	return math.Sqrt(
		math.Pow(p.X-target.X, 2) + math.Pow(p.Y-target.Y, 2),
	)
}

func (p Point) Resonance(f Floorplan) bool {

	var closestPoint Point
	var minCombined = math.MaxFloat64
	var index = 0

	for i, point := range f.pointsRaw {

		diff := p.dist(point)

		if diff < minCombined {
			minCombined = diff
			closestPoint = point
			index = i
		}
	}

	var next = index + 1
	var prev = index - 1
	if index == 0 {
		next = 1
		prev = len(f.pointsRaw) - 2
	}

	distPrevPoint := p.dist(f.pointsRaw[prev])
	distNextPoint := p.dist(f.pointsRaw[next])

	closerDistance := math.Min(distPrevPoint, distNextPoint)
	var closerNeighbor Point
	if closerDistance == distPrevPoint {
		closerNeighbor = f.pointsRaw[prev]
	} else {
		closerNeighbor = f.pointsRaw[next]
	}

	slope := (closestPoint.Y - closerNeighbor.Y) / (closestPoint.X - closerNeighbor.X)
	//equation := (slope * (p.X - closestPoint.X)) + closestPoint.Y
	y := slope*(p.X-closestPoint.X) + closestPoint.Y
	if slope == 0 || slope == -0 {
		y = closestPoint.Y
	}
	x := closestPoint.X + (y / slope) - (closestPoint.Y / slope)
	if slope == 0 || slope == -0 {
		x = closestPoint.X
	}

	xrel := x / p.X
	yrel := y / p.Y

	lbound := 0.90
	ubound := 1.10
	if (xrel > lbound && xrel < ubound) && (yrel > lbound && yrel < ubound) {
		return true
	}

	var r = new(regression.Regression)
	var yr = new(regression.Regression)

	r.SetObserved("x")
	r.SetVar(0, "angle")
	r.SetVar(1, "y")
	r.SetVar(2, "hypotenuse")

	yr.SetObserved("y")
	yr.SetVar(0, "angle")
	yr.SetVar(1, "x")
	yr.SetVar(2, "hypotenuse")

	var edges [][]float64
	var point []float64

	cs := util.ConvertToBoundingBox(closerNeighbor.X, closerNeighbor.Y)
	var boundingbox string
	for _, s := range fmt.Sprintf("%v", cs) {
		switch s {
		case '}', '{':
			continue
		default:
			boundingbox = boundingbox + string(s)
		}
	}

	for _, edge := range strings.Split(boundingbox, " ") {
		e, _ := strconv.ParseFloat(edge, 64)
		point = append(point, e)
		if len(point) == 2 {
			edges = append(edges, point)
			point = []float64{}
		}
	}

	cn := util.ConvertToBoundingBox(closestPoint.X, closestPoint.Y)
	boundingbox = ""
	for _, s := range fmt.Sprintf("%v", cn) {
		switch s {
		case '}', '{':
			continue
		default:
			boundingbox = boundingbox + string(s)
		}
	}
	for _, edge := range strings.Split(boundingbox, " ") {
		e, _ := strconv.ParseFloat(edge, 64)
		point = append(point, e)
		if len(point) == 2 {
			edges = append(edges, point)
			point = []float64{}
		}
	}

	for _, edge := range edges {
		hypotenuse := math.Sqrt(
			math.Pow(edge[0], 2) + math.Pow(edge[1], 2),
		)
		angle := math.Acos(edge[0]/hypotenuse) * 180 / math.Pi
		r.Train(
			regression.DataPoint(edge[0], []float64{math.Abs(angle), edge[1], hypotenuse}),
		)
		yr.Train(
			regression.DataPoint(edge[1], []float64{math.Abs(angle), edge[0], hypotenuse}),
		)
	}

	r.Run()
	yr.Run()

	prediction, err := r.Predict([]float64{math.Abs(p.Angle), p.Y, p.Hypotenuse})
	if err != nil {
		log.Fatal(err)
	}

	yprediction, err := yr.Predict([]float64{math.Abs(p.Angle), p.X, p.Hypotenuse})
	if err != nil {
		log.Fatal(err)
	}

	xrel = math.Abs(prediction / p.X)
	yrel = math.Abs(yprediction / p.Y)

	if (xrel > lbound && xrel < ubound) && (yrel > lbound && yrel < ubound) {
		return true
	}

	d := p.dist(Point{X: prediction, Y: yprediction})

	//log.Println(p.Angle, xrel, yrel, d)

	if d < 0.05 {
		return true
	}

	if p.Angle < -90 {
		yprediction = yprediction * -1
		if yrel < ubound {
			return true
		}
	}

	log.Println(p.Angle, p.X, p.Y)
	log.Println("linearpred", prediction, yprediction)
	log.Println("slopepred", x, y)
	log.Println("predrel", xrel, yrel)

	log.Println(closestPoint, closerNeighbor)
	log.Println(d)

	return false
}

func main() {

	var flat Floorplan
	err := flat.Load("floorplan.csv")
	if err != nil {
		panic(err)
	}

	var writeBuf bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		if scanner.Text() == "#roundtrip" {
			fmt.Println(writeBuf.String())
			writeBuf.Reset()
			continue
		}

		p := strings.Split(scanner.Text(), ",")

		hypotenuse, _ := strconv.ParseFloat(p[0], 64)
		if hypotenuse == 0 {
			continue
		}

		angle, _ := strconv.ParseFloat(p[1], 64)
		x, _ := strconv.ParseFloat(p[4], 64)
		y, _ := strconv.ParseFloat(p[5], 64)
		point := Point{
			Hypotenuse: hypotenuse,
			Angle:      angle,
			X:          x,
			Y:          y,
		}
		if point.Resonance(flat) {
			continue
		}
		// service.messageBuffer <- fmt.Sprintf("%f,%f", point.X, point.Y)
		writeBuf.WriteString(fmt.Sprintf("%f,%f\n", point.X, point.Y))
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
