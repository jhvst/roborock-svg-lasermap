package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"../castle/util"
)

var linking_point bool

var lastBeam util.BeamBoundingBox

func main() {
	var csvbuf bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	// line scanner
	for scanner.Scan() {

		// comment line
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}

		if scanner.Text() == "0	0	0	360	0" {
			continue
		}

		scanner := bufio.NewScanner(strings.NewReader(scanner.Text()))
		scanner.Split(bufio.ScanWords)

		count := 0

		var buf bytes.Buffer

		// row scanner
		for scanner.Scan() {

			if count%360 == 0 {
				fmt.Println("#roundtrip")
			}

			count++
			buf.Write(scanner.Bytes())

			// first field is hypotenuse, second is angle, third is intensity
			if count%3 == 0 {

				fields := bytes.Split(buf.Bytes(), []byte(","))
				hypotenuse, _ := strconv.ParseFloat(string(fields[0]), 64)
				angle, _ := strconv.ParseFloat(string(fields[1]), 64)

				thisBeam, x, y := util.BoundingBox(hypotenuse, angle)
				linking_point = thisBeam.Overlaps(lastBeam)
				lastBeam = thisBeam

				buf.Write([]byte(","))
				buf.WriteString(fmt.Sprint(lastBeam))
				buf.Write([]byte(","))
				buf.Write([]byte(strconv.FormatFloat(x, 'f', -1, 64)))
				buf.Write([]byte(","))
				buf.Write([]byte(strconv.FormatFloat(y, 'f', -1, 64)))

				if linking_point {

					// when we do linking, we want to rollback to the one spot
					// which started it, and also mark it as something we want
					// to preserve -- this makes it easier to filter out the csv
					if !bytes.HasSuffix(csvbuf.Bytes(), []byte(",link\n")) {
						csvbuf.Truncate(strings.LastIndex(csvbuf.String(), "\n"))
						lastLine := strings.LastIndex(csvbuf.String(), "\n")
						fmt.Println(string(csvbuf.Bytes()[lastLine+1:]))
						csvbuf.WriteString(",link\n")
					}

					buf.WriteString(",link")

					// we can avoid single beam points by only deciding to write
					// earlier hits
					fmt.Println(string(buf.Bytes()))
					linking_point = false
				}

				buf.Write([]byte("\n"))
				csvbuf.Write(buf.Bytes())
				buf.Reset()
				continue
			}
			buf.Write([]byte(","))
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading input:", err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	// unrechable code if input is a stream
	angleMap, err := util.Regression(csvbuf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	ur := make(map[float64][]float64)
	lr := make(map[float64][]float64)
	ul := make(map[float64][]float64)
	ll := make(map[float64][]float64)

	for _, v := range angleMap {

		angle := math.Atan(v[1]/v[0]) * 180 / math.Pi
		angle2 := math.Atan(v[3]/v[2]) * 180 / math.Pi

		if v[0] > 0 && v[1] > 0 {
			ur[angle] = []float64{v[0], v[1]}
		}
		if v[0] > 0 && v[1] < 0 {
			lr[angle] = []float64{v[0], v[1]}
		}
		if v[0] < 0 && v[1] > 0 {
			ul[angle] = []float64{v[0], v[1]}
		}
		if v[0] < 0 && v[1] < 0 {
			ll[angle] = []float64{v[0], v[1]}
		}

		if v[2] > 0 && v[3] > 0 {
			ur[angle2] = []float64{v[2], v[3]}
		}
		if v[2] > 0 && v[3] < 0 {
			lr[angle2] = []float64{v[2], v[3]}
		}
		if v[2] < 0 && v[3] > 0 {
			ul[angle2] = []float64{v[2], v[3]}
		}
		if v[2] < 0 && v[3] < 0 {
			ll[angle2] = []float64{v[2], v[3]}
		}
	}

	var lineBuffer bytes.Buffer

	var urk []float64
	for k := range ur {
		urk = append(urk, k)
	}
	sort.Float64s(urk)

	// To perform the opertion you want
	for _, k := range urk {
		lineBuffer.WriteString(fmt.Sprintf("%f,%f\n",
			ur[k][0], ur[k][1],
		))
	}

	var ulk []float64
	for k := range ul {
		ulk = append(ulk, k)
	}
	sort.Float64s(ulk)

	for _, k := range ulk {
		lineBuffer.WriteString(fmt.Sprintf("%f,%f\n",
			ul[k][0], ul[k][1],
		))
	}

	var llk []float64
	for k := range ll {
		llk = append(llk, k)
	}
	sort.Float64s(llk)

	for _, k := range llk {
		lineBuffer.WriteString(fmt.Sprintf("%f,%f\n",
			ll[k][0], ll[k][1],
		))
	}

	var lrk []float64
	for k := range lr {
		lrk = append(lrk, k)
	}
	sort.Float64s(lrk)

	for _, k := range lrk {
		lineBuffer.WriteString(fmt.Sprintf("%f,%f\n",
			lr[k][0], lr[k][1],
		))
	}

	// close the thing
	firstLineIndex := bytes.Index(lineBuffer.Bytes(), []byte("\n"))
	lineBuffer.Write(lineBuffer.Bytes()[:firstLineIndex])

	// when generating initial map
	ioutil.WriteFile("floorplan.csv", bytes.TrimSpace(lineBuffer.Bytes()), 0666)
}
