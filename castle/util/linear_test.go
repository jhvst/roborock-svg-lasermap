package util

import (
	"fmt"
	"log"
	"testing"
)

func TestCode(t *testing.T) {

	fmt.Println(ConvertToBoundingBox(1, 1))

}

func TestSlope(t *testing.T) {

	slope := (-1.186369 - -1.324586) / (1.263478 - 1.244299)
	/*if math.Signbit(p.X) != math.Signbit(p.Y) {
		slope = (math.Abs(closestPoint.Y) - math.Abs(closerNeighbor.Y)) /
			(math.Abs(closestPoint.X) - math.Abs(closerNeighbor.X))
	}*/

	equation := slope*(1.1884-1.244299) + -1.324586
	equation2 := 1.24429 + (-1.45318 / slope) - (-1.324586 / slope)

	log.Println(equation, equation2)
}
