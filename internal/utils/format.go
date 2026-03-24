package utils

import "strconv"

func FormatMoney(amount int) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	s := strconv.FormatInt(int64(amount), 10)
	n := len(s)
	if n <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}

	res := make([]byte, 0, n+(n-1)/3)
	firstGroupLen := n % 3
	if firstGroupLen == 0 {
		firstGroupLen = 3
	}
	res = append(res, s[:firstGroupLen]...)
	for i := firstGroupLen; i < n; i += 3 {
		res = append(res, ',')
		res = append(res, s[i:i+3]...)
	}
	if neg {
		return "-" + string(res)
	}
	return string(res)
}
