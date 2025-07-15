package util

func RGBToInteger(r, g, b int) int {
	return (r << 16) + (g << 8) + b
}
