package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"./util"
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
}
