// 计算2025年木星进入双子座28°的时间
// 运行: go run scripts/jupiter_gemini28.go
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
)

const (
	SE_JUPITER = 5 // Swiss Ephemeris 木星编号
)

func main() {
	// 初始化 Swiss Ephemeris（使用默认路径）
	sweph.Init("./ephe")
	defer sweph.Close()

	// 目标：双子座 28°
	// 双子座 = 第3个星座 (索引2)，起始经度 = 2 * 30 = 60°
	targetLongitude := 60.0 + 28.0 // 双子座 28° = 88°

	fmt.Println("=== 2025年木星进入双子座28°时间计算 ===")
	fmt.Printf("目标经度: %.2f° (双子座28°)\n\n", targetLongitude)

	// 搜索范围：2025年全年
	startJD := sweph.JulDay(2025, 1, 1, 0, true)
	endJD := sweph.JulDay(2026, 1, 1, 0, true)

	// 使用二分查找精确计算木星到达目标经度的时间
	result := findJupiterAtLongitude(targetLongitude, startJD, endJD)

	if result != 0 {
		// 转换回公历日期
		year, month, day, hour := sweph.RevJul(result, true)
		hourInt := int(hour)
		minute := int((hour - float64(hourInt)) * 60)
		second := int(((hour - float64(hourInt)) * 60 - float64(minute)) * 60)

		fmt.Println("✓ 找到精确时间:")
		fmt.Printf("  日期: %04d-%02d-%02d\n", year, month, day)
		fmt.Printf("  时间: %02d:%02d:%02d UTC\n", hourInt, minute, second)

		// 转换为北京时间 (UTC+8)
		beijingTime := time.Date(year, time.Month(month), day, hourInt, minute, second, 0, time.UTC).Add(8 * time.Hour)
		fmt.Printf("  北京时间: %s\n", beijingTime.Format("2006-01-02 15:04:05"))

		// 验证位置
		pos, _ := sweph.CalcUT(result, SE_JUPITER)
		signDeg := pos.Longitude - float64(int(pos.Longitude/30.0))*30.0
		signs := []string{"白羊座", "金牛座", "双子座", "巨蟹座", "狮子座", "处女座",
			"天秤座", "天蝎座", "射手座", "摩羯座", "水瓶座", "双鱼座"}
		signIdx := int(pos.Longitude / 30.0)
		if signIdx < 0 {
			signIdx = 0
		} else if signIdx > 11 {
			signIdx = 11
		}

		fmt.Printf("\n验证位置:\n")
		fmt.Printf("  木星经度: %.4f°\n", pos.Longitude)
		fmt.Printf("  所在星座: %s\n", signs[signIdx])
		fmt.Printf("  星座度数: %.4f°\n", signDeg)
		fmt.Printf("  顺/逆行: %s\n", map[bool]string{true: "逆行", false: "顺行"}[pos.IsRetrograde])
	} else {
		fmt.Println("未找到木星进入双子座28°的时间")
	}
}

// findJupiterAtLongitude 使用二分查找找到木星到达目标经度的精确时间
func findJupiterAtLongitude(targetLon, startJD, endJD float64) float64 {
	// 首先粗略扫描，找到木星经过目标经度的时间段
	step := 1.0 // 每天检查一次

	var prevJD float64
	prevPos, _ := sweph.CalcUT(startJD, SE_JUPITER)
	prevLon := normalizeLongitude(prevPos.Longitude)

	for jd := startJD + step; jd <= endJD; jd += step {
		pos, _ := sweph.CalcUT(jd, SE_JUPITER)
		lon := normalizeLongitude(pos.Longitude)

		// 检查是否经过目标经度（考虑顺行和逆行）
		if hasCrossed(prevLon, lon, targetLon, pos.SpeedLong >= 0) {
			// 在这个区间内进行精确二分查找
			return binarySearchExact(targetLon, prevJD, jd)
		}

		prevJD = jd
		prevLon = lon
	}

	return 0
}

// hasCrossed 检查是否经过了目标经度
func hasCrossed(prevLon, currLon, targetLon float64, isDirect bool) bool {
	if isDirect {
		// 顺行：prevLon < targetLon <= currLon (考虑360度环绕)
		if currLon >= prevLon {
			return prevLon < targetLon && targetLon <= currLon
		}
		// 跨越0度
		return prevLon < targetLon || targetLon <= currLon
	}
	// 逆行：currLon <= targetLon < prevLon (考虑360度环绕)
	if currLon <= prevLon {
		return currLon <= targetLon && targetLon < prevLon
	}
	// 跨越0度
	return currLon <= targetLon || targetLon < prevLon
}

// binarySearchExact 二分查找精确时间
func binarySearchExact(targetLon, startJD, endJD float64) float64 {
	const tolerance = 0.0001 // 精度：约0.01角秒

	low, high := startJD, endJD

	for high-low > 0.0001 { // 约8.6秒精度
		mid := (low + high) / 2
		pos, _ := sweph.CalcUT(mid, SE_JUPITER)
		lon := normalizeLongitude(pos.Longitude)

		diff := angularDifference(lon, targetLon)

		if math.Abs(diff) < tolerance {
			return mid
		}

		// 根据速度方向决定搜索方向
		if pos.SpeedLong >= 0 {
			// 顺行：如果当前经度小于目标，需要更大的JD
			if diff < 0 {
				low = mid
			} else {
				high = mid
			}
		} else {
			// 逆行：如果当前经度小于目标，需要更小的JD
			if diff < 0 {
				high = mid
			} else {
				low = mid
			}
		}
	}

	return (low + high) / 2
}

// normalizeLongitude 将经度归一化到 [0, 360)
func normalizeLongitude(lon float64) float64 {
	lon = math.Mod(lon, 360.0)
	if lon < 0 {
		lon += 360.0
	}
	return lon
}

// angularDifference 计算两个经度的最小差值（考虑360度环绕）
func angularDifference(a, b float64) float64 {
	diff := a - b
	for diff > 180 {
		diff -= 360
	}
	for diff < -180 {
		diff += 360
	}
	return diff
}
