package worldmap

import "math"

func convexHull(pts [][2]float64) [][2]float64 {
	if len(pts) == 0 {
		return nil
	}
	uniq := uniquePoints(pts)
	if len(uniq) == 1 {
		x, y := uniq[0][0], uniq[0][1]
		return [][2]float64{{x - 0.55, y}, {x, y - 0.55}, {x + 0.55, y}, {x, y + 0.55}}
	}
	if len(uniq) == 2 {
		a, b := uniq[0], uniq[1]
		dx, dy := b[0]-a[0], b[1]-a[1]
		lenAB := math.Hypot(dx, dy)
		if lenAB < 1e-6 {
			return convexHull(uniq[:1])
		}
		nx, ny := -dy/lenAB*0.45, dx/lenAB*0.45
		return [][2]float64{
			{a[0] + nx, a[1] + ny},
			{b[0] + nx, b[1] + ny},
			{b[0] - nx, b[1] - ny},
			{a[0] - nx, a[1] - ny},
		}
	}

	// Monotone chain.
	sorted := append([][2]float64(nil), uniq...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j][0] < sorted[i][0] || (sorted[j][0] == sorted[i][0] && sorted[j][1] < sorted[i][1]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	cross := func(o, a, b [2]float64) float64 {
		return (a[0]-o[0])*(b[1]-o[1]) - (a[1]-o[1])*(b[0]-o[0])
	}

	lower := make([][2]float64, 0, len(sorted))
	for _, p := range sorted {
		for len(lower) >= 2 && cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}
	upper := make([][2]float64, 0, len(sorted))
	for i := len(sorted) - 1; i >= 0; i-- {
		p := sorted[i]
		for len(upper) >= 2 && cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}
	hull := append(lower[:len(lower)-1], upper[:len(upper)-1]...)
	return inflateHull(hull, 0.42)
}

func uniquePoints(pts [][2]float64) [][2]float64 {
	out := make([][2]float64, 0, len(pts))
	seen := map[[2]int]bool{}
	for _, p := range pts {
		key := [2]int{int(math.Round(p[0] * 100)), int(math.Round(p[1] * 100))}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}

func inflateHull(hull [][2]float64, pad float64) [][2]float64 {
	if len(hull) == 0 {
		return hull
	}
	cx, cy := 0.0, 0.0
	for _, p := range hull {
		cx += p[0]
		cy += p[1]
	}
	cx /= float64(len(hull))
	cy /= float64(len(hull))
	out := make([][2]float64, len(hull))
	for i, p := range hull {
		dx, dy := p[0]-cx, p[1]-cy
		n := math.Hypot(dx, dy)
		if n < 1e-6 {
			out[i] = [2]float64{p[0] + pad, p[1]}
			continue
		}
		out[i] = [2]float64{p[0] + dx/n*pad, p[1] + dy/n*pad}
	}
	return out
}
