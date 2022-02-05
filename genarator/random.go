package genarator

import (
	"math/rand"
	"time"

	"github.com/google/uuid"

	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomID() string {
	return uuid.New().String()
}

func randomKeyboardLayout() pb.Keyboard_Layout {
	switch rand.Intn(3) {
	case 1:
		return pb.Keyboard_QWERTY
	case 2:
		return pb.Keyboard_QWERTZ
	default:
		return pb.Keyboard_AZERTY
	}
}

func randomGPUBrand() string {
	return randomStringFromSet("Nvidia", "AMD")
}

func randomGPUName(brand string) string {
	switch brand {
	case "AMD":
		return randomStringFromSet(
			"RX 590",
			"RX 580",
			"RX 5700-XT",
			"RX Vega-56",
		)
	default:
		return randomStringFromSet(
			"RTX 2060",
			"RTX 2070",
			"GTX 1660-Ti",
			"GTX 1070",
		)
	}
}

func randomCPUBrand() string {
	return randomStringFromSet("AMD", "Intel")
}

func randomCPUName(brand string) string {
	switch brand {
	case "AMD":
		return randomStringFromSet(
			"Ryzen 7 PRO 2700U",
			"Ryzen 5 PRO 3500U",
			"Ryzen 3 PRO 3200GE",
		)
	default:
		return randomStringFromSet(
			"Xeon E-2286M",
			"Core i9-9980HK",
			"Core i7-9750H",
			"Core i5-9400F",
			"Core i3-1005G1",
		)
	}
}

func randomScreenPanel() pb.Screen_Panel {
	switch rand.Intn(2) {
	case 0:
		return pb.Screen_IPS
	default:
		return pb.Screen_OLED
	}
}

func randomScreenResolution() *pb.Screen_Resolution {
	height := randomIntRange(1080, 4320)
	width := height * 16 / 9

	return &pb.Screen_Resolution{
		Width:  uint32(width),
		Height: uint32(height),
	}
}

func randomLaptopBrand() string {
	return randomStringFromSet("Apple", "Dell", "Lenovo")
}

func randomLaptopName(brand string) string {
	switch brand {
	case "Apple":
		return randomStringFromSet("Macbook Air", "Macbook Pro")
	case "Dell":
		return randomStringFromSet("Latitude", "Vostro", "XPS", "Alienware")
	default:
		return randomStringFromSet("Thinkpad X1", "Thinkpad P1", "Thinkpad P53")
	}
}

func RandomLaptopScore() float64 {
	return float64(randomIntRange(1, 10))
}

// Utils
func randomBool() bool {
	return rand.Intn(2) == 1
}

func randomStringFromSet(set ...string) string {
	if len(set) == 0 {
		return ""
	}

	randomIndex := rand.Intn(len(set))
	return set[randomIndex]
}

func randomIntRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func randomFloat64Range(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func randomFloat32Range(min, max float32) float32 {
	return rand.Float32()*(max-min) + min
}
