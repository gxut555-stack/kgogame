package common

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func GetInt32(s string) int32 {
	i, _ := strconv.Atoi(s)
	return int32(i)
}

// GetLocalIpV4 获取本地ip - 格式ipv4
func GetLocalIpV4() string {
	localIp := "127.0.0.1"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return localIp
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
			//} else if ipnet.IP.To16() != nil {
			//	fmt.Println("IPv6 Address:", ipnet.IP.String())
			//}
		}
	}

	return localIp
}

// 字符串转化成整数切片
// "1,2,3" -->  [1,2,3]
// splitStr表示是逗号，冒号之类
func Strs2Int32Slice(str string, splitStr string) []int32 {
	a := []int32{}
	arr := strings.Split(str, splitStr)
	for _, v := range arr {
		i, _ := strconv.Atoi(v)
		a = append(a, int32(i))
	}
	return a
}

// [1, 2, 3] --> "1, 2, 3"
func Int32Slice2Str(a []int32) string {
	str := ""
	for _, v := range a {
		if str == "" {
			str += fmt.Sprintf("%d", v)
			continue
		}
		str += fmt.Sprintf(",%d", v)
	}
	return str
}

func Md5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func Md5ByJson(data any) string {
	byt, _ := json.Marshal(data)
	return Md5(byt)
}

func MoneyToFloatString(money int64) string {
	var minus bool
	if money == 0 {
		return "0"
	} else if money < 0 {
		money = -money
		minus = true
	}
	if minus {
		return fmt.Sprintf("-%d.%d", money/1000, money%1000)
	}
	return fmt.Sprintf("%d.%d", money/1000, money%1000)
}

// 定义要去除的常见前缀
var prefixesToRemove = []string{
	"player-", "bingoplus", "arenaplus", "gzone",
}

// 去除名字中的特定前缀
func removePrefixes(name string) string {
	lowerName := strings.ToLower(name)
	for _, prefix := range prefixesToRemove {
		if strings.HasPrefix(lowerName, prefix) {
			// 获取前缀长度
			prefixLen := len(prefix)
			// 保留前缀之后的部分
			return strings.TrimSpace(lowerName[prefixLen:])
		}
	}
	return lowerName
}

func levenshteinDistance(s1, s2 string) int {
	// 去除前缀
	s := removePrefixes(s1)
	t := removePrefixes(s2)
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}

	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]     // 删除
				if d[i][j-1] < min { // 插入
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min { // 替换
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}
	}
	return d[len(s)][len(t)]
}

// 计算相似度（0-1之间）
func StringSimilarity(s1, s2 string) float64 {
	maxLen := strMax(len(s1), len(s2))
	if maxLen == 0 {
		return 1.0
	}
	distance := levenshteinDistance(s1, s2)
	return 1.0 - float64(distance)/float64(maxLen)
}
func strMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 字符串格式转换成数字类型
func StringToNumber[T int32 | int64 | int | uint8 | uint32 | int16](s string) (T, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("string=%s to integer err: %v", s, err)
	}
	return T(n), nil
}

// 不区分大小写判断是否包含子字符串
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}
