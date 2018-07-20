package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"../castle/util"
)

func scan(points [][]float64) ([][]float64, bool) {
	var matched bool
	for i, point := range points {
		if len(point) == 0 {
			continue
		}
		log.Println(point)
		log.Println(points)
		this := util.ConvertToBoundingBox(point[0], point[1])
		if i == 0 {
			if this.Overlaps(
				util.ConvertToBoundingBox(
					points[len(points)-1][0],
					points[len(points)-1][1],
				),
			) {
				points[0] = append(points[0],
					points[len(points)-1]...,
				)
				points[len(points)-1] = []float64{}
				matched = true
				continue
			}
		}
		if i == len(points)-1 {
			continue
		}

		if len(points[i+1]) == 0 {
			log.Println("problem")
			continue
		}

		if this.Overlaps(
			util.ConvertToBoundingBox(
				points[i+1][0],
				points[i+1][1],
			),
		) {
			points[i] = append(points[i],
				points[i+1]...,
			)
			points[i+1] = []float64{}
			matched = true
			continue
		}

		log.Println("second")

		if this.Overlaps(
			util.ConvertToBoundingBox(
				points[i+1][len(points[i+1])-2],
				points[i+1][len(points[i+1])-1],
			),
		) {
			points[i] = append(points[i],
				points[i+1]...,
			)
			points[i+1] = []float64{}
			matched = true
			continue
		}
	}
	return points, matched
}

func recurs(points [][]float64) [][]float64 {
	points, m := scan(points)
	var nonEmpty [][]float64
	for _, point := range points {
		if len(point) == 0 {
			continue
		}
		nonEmpty = append(nonEmpty, point)
	}
	if m {
		return recurs(nonEmpty)
	}
	return nonEmpty
}

type Point struct {
	X, Y       float64
	Hypotenuse float64
	Angle      float64
}

func (p Point) dist(target Point) float64 {
	return math.Sqrt(
		math.Pow(p.X-target.X, 2) + math.Pow(p.Y-target.Y, 2),
	)
}

func handler(w http.ResponseWriter, r *http.Request) {
	parseLog()
}

func parseLog() {
	data, err := ioutil.ReadFile("../classifier/stream.csv")
	if err != nil {
		log.Fatal(err)
	}
	data = bytes.TrimSpace(data)
	//data = bytes.Replace(data, []byte("\n\n\n"), []byte("\n\n"), -1)

	lidarSpins := strings.Split(string(data), "\n\n")

	var buf bytes.Buffer
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9} {
		buf.WriteString(lidarSpins[len(lidarSpins)-v] + "\n")
	}

	log.Println(buf.String())

	var points [][]float64
	scanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	for scanner.Scan() {

		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}

		values := strings.Split(scanner.Text(), ",")

		x, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			panic(err)
		}

		y, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			panic(err)
		}

		points = append(points, []float64{x, y})
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	angleMap := make(map[float64][]float64)
	for _, point := range points {
		hypotenuse := math.Sqrt(
			math.Pow(point[0], 2) + math.Pow(point[1], 2),
		)
		angle := math.Acos(point[0]/hypotenuse) * 180 / math.Pi
		angleMap[angle] = append(angleMap[angle], point...)
	}

	var keys []float64
	for k := range angleMap {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	// To perform the opertion you want
	var pointsOrdered []Point
	for _, k := range keys {
		pointsOrdered = append(pointsOrdered, Point{
			X: angleMap[k][0],
			Y: angleMap[k][1],
		})
	}

	var objects [][]Point
	for i, p := range pointsOrdered {

		if i == len(pointsOrdered)-1 {
			continue
		}

		if p.dist(pointsOrdered[i+1]) < 0.10 {

			if len(objects) == 0 {
				objects = append(objects, []Point{p})
			}

			lastObject := len(objects) - 1
			objects[lastObject] = append(objects[lastObject], p)
			continue
		}

		if len(objects) == 0 {
			objects = append(objects, []Point{p})
		}

		// if broken, create new object
		objects = append(objects, []Point{pointsOrdered[i+1]})
	}

	biggestObject := len(objects[0])
	biggestObjectIndex := 0
	for i, object := range objects {
		if i == 0 {
			continue
		}
		if len(object) > biggestObject {
			biggestObject = len(object)
			biggestObjectIndex = i
		}
	}

	points = [][]float64{}
	for _, point := range objects[biggestObjectIndex] {
		points = append(points, []float64{point.X, point.Y})
	}

	hypotenuse := math.Sqrt(
		math.Pow(points[0][0], 2) + math.Pow(points[0][1], 2),
	)
	angle := math.Acos(points[0][0]/hypotenuse) * 180 / math.Pi

	if math.Signbit(points[0][0]) && math.Signbit(points[0][1]) {
		angle = angle * -1
	}

	d := objects[biggestObjectIndex][0].dist(
		objects[biggestObjectIndex][len(objects[biggestObjectIndex])-1],
	) * 100 * 2 // m to cm

	conf := math.Min(d, 8) / math.Max(d, 8) * 100

	result := fmt.Sprintf("Statement: With a confidence of %f percent, human leg detected %.1f meters away, at an angle of %.1f degrees.", conf, hypotenuse, angle)

	log.Println(result)

	// TTS command on Mac
	// cmd := exec.Command("say", result)
	// err = cmd.Start()
	// if err != nil {
	// 		log.Fatal(err)
	// }
}

// run on command, should go through the stream csv and
// analyze it, can read into memory
func main() {
	http.HandleFunc("/tail", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
