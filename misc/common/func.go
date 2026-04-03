package common

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"kgogame/util/gconf"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const nonstr = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
const lowernonstr = "abcdefghjkmnpqrstuvwxyz23456789"

var chars = map[int]string{
	0:  "a",
	1:  "b",
	2:  "c",
	3:  "d",
	4:  "e",
	5:  "f",
	6:  "g",
	7:  "h",
	8:  "i",
	9:  "j",
	10: "k",
	11: "l",
	12: "m",
	13: "n",
	14: "o",
	15: "p",
	16: "q",
	17: "r",
	18: "s",
	19: "t",
	20: "u",
	21: "v",
	22: "w",
	23: "x",
	24: "y",
	25: "z",
}

func init() {
	rand.Seed(time.Now().Unix() * int64(os.Getpid()))
}

// 数字转字符串
func Convert2str(n int) string {
	ret := ""
	for {
		mod := n % 26
		ret = chars[mod] + ret
		n -= mod
		if n == 0 {
			break
		} else {
			n /= 26
		}
	}
	return ret
}

// 检查某个时间戳是否是今天
func IsToday(timestamp int64) bool {
	y, m, d := time.Now().Date()
	y2, m2, d2 := time.Unix(timestamp, 0).Date()
	return y == y2 && m == m2 && d == d2
}

// 检查是否是同一天
func IsSameDay(t1, t2 int64) bool {
	y1, m1, d1 := time.Unix(t1, 0).Date()
	y2, m2, d2 := time.Unix(t2, 0).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsSameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// 获取某天0点的时间
func GetZeroTime(t time.Time) time.Time {
	//loc, _ := time.LoadLocation("Asia/Kolkata")
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return tm1
}

// 获取某天23点59分59秒的时间
func GetEndTime(t time.Time) time.Time {
	//loc, _ := time.LoadLocation("Asia/Kolkata")
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
	return tm1
}

// 获取某一天的小时时间
func GetHourTime(t time.Time, hour int) time.Time {
	//loc, _ := time.LoadLocation("Asia/Kolkata")
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, time.Local)
	return tm1
}

// 获取某一天的分钟
func GetMinuteTime(t time.Time) time.Time {
	//loc, _ := time.LoadLocation("Asia/Kolkata")
	tm1 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
	return tm1
}

// 获取某月份1日0点
func GetMonthStartTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
}

// 获取某月份最后一天23点59分59秒
func GetMonthEndTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local).
		AddDate(0, 1, 0).
		Add(-time.Second)
}

// 计算两个时间相差几天，t1 > t2
// return 0同一天
func TimeSub(t1, t2 time.Time) int32 {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.Local)
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.Local)

	return int32(t1.Sub(t2).Hours() / 24)
}

// 转换成年月日
func ToYmd(unix int64, day int32) int32 {
	//loc, _ := time.LoadLocation("Asia/Kolkata")
	t := time.Unix(unix+int64(day)*86400, 0).Format("20060102")
	v, _ := strconv.Atoi(t)
	return int32(v)
}

// 获取本周一0点
func GetFirstDateOfWeek() time.Time {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	//loc, _ := time.LoadLocation("Asia/Kolkata")
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)

	return weekStartDate
}

/**
 * 生成随机数
 */
func GetRandInt32(min, max int32) int32 {
	rand.Seed(time.Now().UnixNano() + rand.Int63n(9999))
	if max-min == 0 {
		return min
	}
	return rand.Int31n(max-min) + min
}

// 检查邮箱是否正确
func IsEmail(email string) bool {
	isMatch, _ := regexp.MatchString(`^([a-zA-Z0-9_\.-]+)@([\dA-Za-z\.-]+)\.([A-Za-z\.]{2,})$`, email)
	if isMatch {
		return true
	}
	return false
}

// 检测upi账号是否正确
func IsUpiAccount(account string) bool {
	ret, _ := regexp.MatchString(`^([a-zA-Z0-9]+)@([A-Za-z]+)$`, account)
	return ret
}

func IsBankAccount(account string) bool {
	ret, _ := regexp.MatchString(`^([0-9]{5,35})$`, account)
	return ret
}

// 检测持卡人名字是否正确
func IsHolderName(account string) bool {
	ret, _ := regexp.MatchString(`^[a-zA-Z]([a-zA-Z ]{2,48})[a-zA-Z]$`, account)
	return ret
}

// md5加密密码
func ConvertPwd(pwd string) string {
	ret := md5.Sum([]byte(fmt.Sprintf("^YYSK|%s|PWD$", pwd)))
	return fmt.Sprintf("%x", ret)
}

// 是否中文
func IsChinese(name string) bool {
	if name == "" {
		return false
	}
	for _, v := range name {
		if unicode.Is(unicode.Han, v) {
			return true
		}
	}
	return false
}

// 将字符串格式的版本号转为数字格式
func VerStr2Num(ver string) int64 {
	num, parts, times := int64(0), strings.SplitN(ver, ".", 4), []int64{1000000000, 1000000, 1000, 1}
	for k, v := range parts {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n < 1000 {
			num += int64(n) * times[k]
		}
	}
	return num
}

// 将数字格式的版本号转为字符串格式
func VerNum2Str(num int64) string {
	parts := []int64{0, 0, 0, 0}
	if num >= 1000000000 {
		parts[0] = num / 1000000000
	}
	if num >= 1000000000 {
		parts[1] = (num % 1000000000) / 1000000
	}
	if num >= 1000000000 {
		parts[2] = (num % 1000000) / 1000
	}
	parts[3] = num % 1000
	return fmt.Sprintf("%d.%d.%d.%d", parts[0], parts[1], parts[2], parts[3])
}

// 版本比较，ver1大于ver2返回正数，ver1等于ver2返回0，ver1小于ver2返回负数
func CompareVersion(ver1 string, ver2 string) int64 {
	return VerStr2Num(ver1) - VerStr2Num(ver2)
}

// 版本比较，ver1大于ver2返回正数，ver1等于ver2返回0，ver1小于ver2返回负数
// 多米诺应用版本号和德州应用版本号不一致，所以多米诺先不限制版本号
func CompareVersion2(ver1 string, ver2 string, platform int32) int64 {
	return VerStr2Num(ver1) - VerStr2Num(ver2)
}

func GenRandCode(num int) string {
	authCode := ""
	for i := 0; i < num; i++ {
		authCode += randCode()
	}
	return authCode
}

func randCode() string {
	s := rand.Intn(len(nonstr))
	return nonstr[s : s+1]
}

func GenRandLowerCode(num int) string {
	authCode := ""
	for i := 0; i < num; i++ {
		authCode += randLowerCode()
	}
	return authCode
}

func randLowerCode() string {
	s := rand.Intn(len(lowernonstr))
	return lowernonstr[s : s+1]
}

// int64 to string slice
func Int64SliceToStringSlice(info []int64) []string {
	if len(info) < 0 {
		return nil
	}
	var strSlice []string
	for _, v := range info {
		str := strconv.Itoa(int(v))
		strSlice = append(strSlice, str)
	}
	return strSlice
}

// 将字符串切片转换为数字类型切片
func StringSliceToIntSlice[T Number](v []string) []T {
	if len(v) == 0 {
		return nil
	}
	var res []T
	for _, s := range v {
		res = append(res, StringToInt[T](s))
	}
	return res
}

// 字符串格式转换成数字类型
func StringToInt[T Number](s string) T {
	var t T
	switch any(t).(type) {
	case int32:
		n, _ := strconv.ParseInt(s, 10, 32)
		return T(n)
	case int64:
		n, _ := strconv.ParseInt(s, 10, 64)
		return T(n)
	case float64:
		n, _ := strconv.ParseFloat(s, 64)
		return T(n)
	case float32:
		n, _ := strconv.ParseFloat(s, 32)
		return T(n)
	default:
		n, _ := strconv.Atoi(s)
		return T(n)
	}
}

// 字符串格式转换
func Int2Str(num int64, dec int) string {
	val, unit := float64(0), ""
	if num < 1000 {
		val, unit = float64(num), ""
	} else if num < 1000000 {
		val, unit = float64(num)/1000, "K"
	} else if num < 1000000000 {
		val, unit = float64(num)/1000000, "M"
	} else {
		val, unit = float64(num)/1000000000, "B"
	}
	switch dec {
	case 1:
		return fmt.Sprintf("%.1f%s", val, unit)
	case 2:
		return fmt.Sprintf("%.2f%s", val, unit)
	case 3:
		return fmt.Sprintf("%.3f%s", val, unit)
	default:
		return fmt.Sprintf("%d%s", int64(val), unit)
	}
}

// 整形格式化
func IntFormat(val int64) string {
	str := fmt.Sprintf("%d", val)
	if n := len(str); n <= 3 {
		return fmt.Sprintf("%d", val)
	} else if n%3 != 0 {
		str = strings.Repeat(" ", 3-(n%3)) + str
	}
	fields := len(str) / 3
	arr := make([]string, fields)
	for i := 0; i < fields; i++ {
		start, end := i*3, (i+1)*3
		arr[i] = str[start:end]
	}
	return strings.TrimLeft(strings.Join(arr, ","), " ")
}

// money 格式化
func MoneyFormat(num int64) string {
	var kilo int64 = 1000
	var million = kilo * kilo
	var billion = kilo * kilo * kilo
	var trillion = kilo * kilo * kilo * kilo

	var unitStr string
	var level = kilo

	if num < kilo {
		return fmt.Sprintf("%d", num)
	} else if num < million {
		unitStr = "K"
		level = kilo
	} else if num < billion {
		unitStr = "M"
		level = million
	} else if num < trillion {
		unitStr = "B"
		level = billion
	} else {
		unitStr = "T"
		level = trillion
	}

	intNum := int(math.Floor(float64(num) / float64(level)))
	point1 := int(math.Floor(float64(num%level) / float64(level/10)))
	point2 := int(math.Floor(float64(num%(level/10)) / float64(level/100)))

	if point1 == 0 && point2 == 0 {
		return fmt.Sprintf("%d%s", intNum, unitStr)
	} else if point2 == 0 {
		return fmt.Sprintf("%d.%d%s", intNum, point1, unitStr)
	} else {
		return fmt.Sprintf("%d.%d%d%s", intNum, point1, point2, unitStr)
	}
}

// 数组打乱
func ShuffleInt64(arr []int64) []int64 {
	var length = len(arr)
	if length == 0 {
		return arr
	}

	rand.Seed(time.Now().UnixNano())
	for i := length - 1; i >= 0; i-- {
		n := rand.Intn(i + 1)
		arr[n], arr[i] = arr[i], arr[n]
	}

	return arr
}
func ShuffleInt32(arr []int32) []int32 {
	var length = len(arr)
	if length == 0 {
		return arr
	}

	rand.Seed(time.Now().UnixNano())
	for i := length - 1; i >= 0; i-- {
		n := rand.Intn(i + 1)
		arr[n], arr[i] = arr[i], arr[n]
	}

	return arr
}

// 数字转IP
func Long2Ip(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

// IP转数字
func Ip2Long(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

// 取指定时间的当日零点的秒数
func SecondsToDayStart(t int64) int64 {
	bTime := time.Unix(t, 0)
	return time.Date(bTime.Year(), bTime.Month(), bTime.Day(), 0, 0, 0, 0, time.Local).Unix()
}

// 取指定时间的当日23:59:59的秒数
func SecondsToDayEnd(t int64) int64 {
	return SecondsToDayStart(t) + 86400 - 1
}

// 到明天零点的秒数
func SecondsToTomorrow() int64 {
	now := time.Now()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	return today.Unix() + 86400 - now.Unix()
}

// 日期转时间戳
func ToUnix(date int32) int64 {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	dateStr := strconv.Itoa(int(date))
	dateTime, _ := time.ParseInLocation("20060102", dateStr, loc)

	return dateTime.Unix()
}

// 转换安全字符串
func ToSafeStr(msg string, maxLen int) string {
	msgtxt := ([]rune)(msg)
	if len(msgtxt) > maxLen {
		msgtxt = msgtxt[0:maxLen]
	}
	return strings.Replace(strings.Replace(string(msgtxt), "<", "＜", -1), ">", "＞", -1)
}

func InArray32(arr []int32, num int32) bool {
	//fast try
	if len(arr) == 1 {
		return num == arr[0]
	} else if len(arr) == 2 {
		return num == arr[0] || num == arr[1]
	} else if len(arr) == 3 {
		return num == arr[0] || num == arr[1] || num == arr[2]
	} else if len(arr) == 0 {
		return false
	}
	for _, v := range arr {
		if v == num {
			return true
		}
	}

	return false
}

func InArray64(arr []int64, num int64) bool {
	for _, v := range arr {
		if v == num {
			return true
		}
	}

	return false
}

// 泛型
func InArray[T comparable](arr []T, num T) bool {
	//fast try
	if len(arr) == 1 {
		return num == arr[0]
	} else if len(arr) == 2 {
		return num == arr[0] || num == arr[1]
	} else if len(arr) == 3 {
		return num == arr[0] || num == arr[1] || num == arr[2]
	} else if len(arr) == 0 {
		return false
	}
	for _, v := range arr {
		if v == num {
			return true
		}
	}

	return false
}

func Join[T Number](arr []T, sep string) string {
	str := ""
	for i := range arr {
		str += strconv.Itoa(int(arr[i])) + sep
	}
	return str[:len(str)-1]
}

// 字符串数组查找
func InStringArray(arr []string, str string) bool {
	//fast try
	if len(arr) == 1 {
		return str == arr[0]
	} else if len(arr) == 2 {
		return str == arr[0] || str == arr[1]
	} else if len(arr) == 3 {
		return str == arr[0] || str == arr[1] || str == arr[2]
	} else if len(arr) == 0 {
		return false
	}
	for _, v := range arr {
		if v == str {
			return true
		}
	}

	return false
}

// RemoveDuplicateSliceString slice(string类型)元素去重
func RemoveDuplicateSliceString(slc []string) []string {
	if len(slc) <= 1 {
		return slc
	}
	result := make([]string, 0, len(slc))
	temp := map[string]struct{}{}
	for _, item := range slc {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// RemoveDuplicateSliceInt32 slice(int32类型)元素去重
func RemoveDuplicateSliceInt32(slc []int32) []int32 {
	if len(slc) <= 1 {
		return slc
	}
	result := make([]int32, 0, len(slc))
	temp := map[int32]struct{}{}
	for _, item := range slc {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// 是否正式环境
func IsProduct() bool {
	return gconf.GetStringConf("env") == "pro"
}

func IsBrazil() bool {
	return gconf.GetStringConf("project") == "BR"
}

// 是否测试环境
func IsTestEnv() bool {
	return gconf.GetStringConf("env") == "test"
}

// 是否测试环境
func IsDevEnv() bool {
	return gconf.GetStringConf("env") == "dev"
}

func GetEnv() string {
	return gconf.GetStringConf("env")
}

// 获取环境名称
func GetEnvName() string {
	env := GetEnv()
	switch env {
	case "pro":
		return "正式环境"
	case "test":
		return "测试环境/UAT"
	case "dev":
		return "开发环境"
	default:
		return fmt.Sprintf("未知环境-%s", env)
	}
}

// 提现方式转换成名字
func ToBankCode(bankCode int32) string {
	switch bankCode {
	case 100:
		return "银行"
	case 103:
		return "UPI"
	}

	return "-"
}

func ABS(val int64) int64 {
	if val < 0 {
		return -val
	} else {
		return val
	}
}

// ToRupee 金币转换成卢比
func ToRupee(amount int64) int64 {
	return amount / 1000
}

func ToRupeeFloat(amount int64) float64 {
	return float64(amount) / 1000
}

// 获取现在是1年中的第几周
func GetWeek(t time.Time) int32 {
	_, week := t.ISOWeek()
	return int32(week)
}

// 获取某周的开始和结束时间,week为0本周,-1上周，1下周以此类推
func WeekIntervalTime(week int) (startTime time.Time, endTime time.Time) {
	now := time.Now()
	offset := int(time.Monday - now.Weekday())
	//周日做特殊判断 因为time.Monday = 0
	if offset > 0 {
		offset = -6
	}

	year, month, day := now.Date()
	thisWeek := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	startTime = thisWeek.AddDate(0, 0, offset+7*week)
	endTime = thisWeek.AddDate(0, 0, offset+6+7*week)

	return startTime, endTime
}

// MergeArray 求两个切片的并集
func MergeArray[T Number](a []T, b []T) []T {
	var mergeArray []T
	temp := map[T]struct{}{}

	for _, val := range b {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
			mergeArray = append(mergeArray, val)
		}
	}

	for _, val := range a {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
			mergeArray = append(mergeArray, val)
		}
	}

	return mergeArray
}

// A移除B
func ArrARemoveB[T Number](A []T, B []T) []T {
	aNew := make([]T, 0, len(A))
	eleCountB := make(map[T]int)
	for _, ele := range B {
		eleCountB[ele]++
	}
	for _, ele := range A {
		if eleCountB[ele] > 0 {
			eleCountB[ele]--
		} else {
			aNew = append(aNew, ele)
		}
	}
	return aNew
}

// A是B的子集
func ArrAIsSubB[T Number](A []T, B []T) bool {
	eleCountB := make(map[T]int)
	for _, ele := range B {
		eleCountB[ele]++
	}
	for _, ele := range A {
		if eleCountB[ele] == 0 {
			return false
		}
		eleCountB[ele]--
	}
	return true
}

// DiffArray 求A 不在B 差异值
func DiffArray(a []int32, b []int32) []int32 {
	var diffArray []int32
	temp := map[int32]struct{}{}

	for _, val := range a {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
		}
	}

	for _, val := range b {
		if _, ok := temp[val]; ok {
			delete(temp, val)
		}
	}
	for i, _ := range temp {
		diffArray = append(diffArray, i)
	}
	return diffArray
}

// 求A和B交集
func UniteArray[T comparable](a []T, b []T) []T {
	var unite []T
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			if b[j] == a[i] {
				unite = append(unite, a[i])
				break
			}
		}
	}
	return unite
}

func IndexArray[S ~[]E, E comparable](s S, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

func EqualArray(s1, s2 []int32) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func DeleteFromArray[S ~[]E, E comparable](s S, v ...E) S {
	for _, item := range v {
		i := IndexArray(s, item)
		if i < 0 {
			continue
		}
		s = append(s[:i], s[i+1:]...)
	}
	return s
}

// 生成在区间 [a, b] 中的随机整数
func RandInt(a, b int) int {
	if a > b {
		return b + rand.Int()%(a-b+1)
	} else {
		return a + rand.Int()%(b-a+1)
	}
}

// 生成在区间 [a, b] 中的随机整数
func RandInt32(a, b int32) int32 {
	if a > b {
		return b + rand.Int31()%(a-b+1)
	} else {
		return a + rand.Int31()%(b-a+1)
	}
}

func InArrayString(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}

	return false
}

func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		// or通过ASCII码进行大小写的转化
		// 65-90（A-Z），97-122（a-z）
		//判断如果字母为大写的A-Z就在前面拼接一个_
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	//ToLower把大写字母统一转小写
	return strings.ToLower(string(data[:]))
}

// 生成在区间 [a, b] 中的随机整数
func RandInt64(a, b int64) int64 {
	if a > b {
		return b + rand.Int63()%(a-b+1)
	} else {
		return a + rand.Int63()%(b-a+1)
	}
}

// 获取北京时间
func GetBeiJingDate() string {
	now := time.Now()
	if gconf.IsBRProject() {
		//巴西加11小时
		now = now.Add(time.Hour * time.Duration(11))
	} else {
		//印度加2个半小时
		now = now.Add(time.Minute * time.Duration(150))
	}

	return now.Format("2006-01-02 15:04:05")
}

// IsAlphaNumeric 判断字符串是否由字母数字组成
func IsAlphaNumeric(str string) bool {
	// 遍历字符串，判断每个字符是否为字母数字
	for _, ch := range str {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}

type Number interface {
	int | int64 | int32 | int16 | int8 | uint64 | uint32 | uint16 | uint8 | uint | float64 | float32
}

// 数字类型比较大小
func GenericMin[T Number](a, b T) T {
	if a > b {
		return b
	}
	return a
}

// 数字类型比较大小
func GenericMax[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// 获取http 远端 ip
func GetHttpClientIp(r *http.Request) string {
	// 这里也可以通过X-Forwarded-For请求头的第一个值作为用户的ip
	// 但是要注意的是这两个请求头代表的ip都有可能是伪造的
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		if ip = r.Header.Get("X-Forwarded-For"); ip == "" {
			// 当请求头不存在即不存在代理时直接获取ip
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}
	}
	return ip
}

// playlog是否用新服务
func IsUseNewServer(t time.Time) bool {
	if IsProduct() {
		return t.Unix() >= 1703606400 // 2024/12-27
	} else {
		return t.Unix() >= 1702310400 // 2023-12-12 1702310400
	}
}

// gamelog读写是否用新服务
func IsUseNewGameServer(t time.Time) bool {
	if IsProduct() {
		return t.Unix() >= 1703779200 // 2023-12-29
	} else {
		return t.Unix() >= 1703779200 // 2023-12-29
	}
}

// gamelog读写是否用新服务
func IsUseMongoByGameLog(t time.Time) bool {
	if IsProduct() {
		return t.Unix() >= 1720843200 // 2024-07-13 12:00:00
	} else {
		return t.Unix() >= 1720540800 // 2024-07-10
	}
}

// gamelog读ID是否用新服务
func IsUseMongoByAddGameLog(t time.Time) bool {
	if IsProduct() {
		return t.Unix() >= 1719590400 // 2024-06-29
	} else {
		return t.Unix() >= 1719590400 // 2024-06-29
	}
}

// gamelog读写是否用分表
func IsUseNewGameSubfix(t time.Time) bool {

	if t.Unix() >= 1714884008 && t.Unix() <= 1714918322 { //线上bug兼容 2024/5/5 12:40:08  到  2024/5/5 22:12:02 读主表数据
		return false
	}

	if IsProduct() {
		return t.Unix() >= 1714129200 // 2024-04-26 19:00:00
	} else {
		return t.Unix() >= 1714110598 // 2024-04-26 13:49:58
	}
}

func PowInt(base, exponent int) int {
	result := 1
	for i := 0; i < exponent; i++ {
		result *= base
	}
	return result
}

func PowInt32(base, exponent int32) int32 {
	result := int32(1)
	for i := int32(0); i < exponent; i++ {
		result *= base
	}
	return result
}

// 指定数量切分
func ChunkSlice[T Number](slice []T, chunkSize int) [][]T {
	var result [][]T

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		result = append(result, slice[i:end])
	}

	return result
}

func MustNil(err error) {
	if err != nil {
		panic(err)
	}
}

// 切分为数字数组
func Split2Number[T Number](s, sep string) ([]T, error) {
	if len(s) == 0 {
		return make([]T, 0), nil
	}

	sl := strings.Split(s, sep)
	var res = make([]T, 0, len(sl))
	for _, v := range sl {
		i64, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseInt %s failed %s", v, err)
		}

		res = append(res, T(i64))
	}
	return res, nil
}

// url encode alias url.QueryEscape
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}

// url decode alias url.QueryUnescape
func UrlDecode(str string) (string, error) {
	return url.QueryUnescape(str)
}

// 将json字符串转换为int数组
func JsonStr2IntArray[T Number](s string) ([]T, error) {
	if len(s) == 0 {
		return make([]T, 0), nil
	}

	var res = make([]T, 0, 10)
	if err := json.Unmarshal([]byte(s), &res); err != nil {
		return nil, fmt.Errorf("json.Unmarshal %s failed %s", s, err)
	}
	return res, nil
}

// 数组格式转换成字符串，sep: 分隔符
func IntSliceJoin2String[T Number](arr []T, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	var str string
	for _, v := range arr {
		str += strconv.Itoa(int(v)) + sep
	}
	return strings.TrimRight(str, sep)
}

// 判断字符串是否为数字
func IsNumeric(str string) bool {
	// 遍历字符串，判断每个字符是否为数字
	for _, ch := range str {
		if !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}

type JsonLog struct {
	val interface{}
}

func (jl *JsonLog) String() string {
	if jl == nil {
		return ""
	}
	v, err := json.Marshal(jl.val)
	if err != nil {
		return fmt.Sprintf("json.encode error %s", err.Error())
	} else {
		return UnsafeBytesToString(v)
	}
}

// 结构体日志输出json字符串
func Struct2JsonLog(val interface{}) *JsonLog {
	if val == nil {
		return nil
	}
	return &JsonLog{val: val}
}
