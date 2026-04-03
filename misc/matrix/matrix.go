package matrix

import (
	"errors"
	"fmt"
	"kgogame/misc/common"
	"strconv"
	"strings"
)

var (
	ErrInvalidData = errors.New("matrix: invalid data")
)

// 位置信息
type MatrixPosition struct {
	Row int
	Col int
}

type Matrix[T common.Number] struct {
	rows int   //行数目
	cols int   //列数目
	data [][]T //矩阵
}

func CreateMatrix[T common.Number](row, col int) *Matrix[T] {
	m := &Matrix[T]{
		cols: col,
		rows: row,
		data: make([][]T, row),
	}
	for i := range m.data {
		m.data[i] = make([]T, col)
	}
	m.Reset()
	return m
}

func (m *Matrix[T]) Rows() int {
	return m.rows
}

func (m *Matrix[T]) Cols() int {
	return m.cols
}

func (m *Matrix[T]) Set(row, col int, val T) {
	if row < 0 || row >= m.rows {
		return
	}
	if col < 0 || col >= m.cols {
		return
	}
	m.data[row][col] = val
}

func (m *Matrix[T]) Get(row, col int) T {
	if row < 0 || row >= m.rows {
		return 0
	}
	if col < 0 || col >= m.cols {
		return 0
	}
	return m.data[row][col]
}

func (m *Matrix[T]) GetRowSymbols(row int) (sl []T) {
	for i := 0; i < m.cols; i++ {
		sl = append(sl, m.data[row][i])
	}
	return sl
}

func (m *Matrix[T]) GetColSymbols(col int) (sl []T) {
	for i := 0; i < m.rows; i++ {
		sl = append(sl, m.data[i][col])
	}
	return sl
}

// 交换值
func (m *Matrix[T]) Swap(row1, col1, row2, col2 int) {
	m.data[row1][col1], m.data[row2][col2] = m.data[row2][col2], m.data[row1][col1]
}

func (m *Matrix[T]) Reset() {
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			m.data[i][j] = 0
		}
	}
}

func (m *Matrix[T]) Clone() *Matrix[T] {
	c := CreateMatrix[T](m.rows, m.cols)
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			c.data[i][j] = m.data[i][j]
		}
	}
	return c
}

// copy
func (m *Matrix[T]) Copy(src *Matrix[T]) {
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			m.data[i][j] = src.data[i][j]
		}
	}
}

func (m *Matrix[T]) Debug() string {
	str := ""
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			str += fmt.Sprintf("m[%d][%d]=%d |", i, j, m.data[i][j])
		}
		str += "\n"
	}
	return str
}

// 返回格式 row-col-val|row-col-val|...
func (m *Matrix[T]) String() string {
	str := ""
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if i == 0 && j == 0 {
				str += fmt.Sprintf("%d-%d-%d", i, j, m.data[i][j])
			} else {
				str += fmt.Sprintf("|%d-%d-%d", i, j, m.data[i][j])
			}
		}
	}
	return str
}

func (m *Matrix[T]) PositionSymbol(list interface{}, create func(row, col int, val T, list interface{})) {
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			create(i, j, m.data[i][j], list)
		}
	}
}

type IPositionSymbol interface {
	GetRow() int32
	GetCol() int32
}

func (m *Matrix[T]) PositionSymbol2Matrix(list []IPositionSymbol, getSymbolFn func(key int) T) {
	for k, v := range list {
		m.Set(int(v.GetRow()), int(v.GetCol()), getSymbolFn(k))
	}
}

func (m *Matrix[T]) FillWithString(s string, symbolFn func(string) T) {
	sl := strings.Split(s, "|")
	for _, v := range sl {
		sl1 := strings.Split(v, "-")
		if len(sl1) == 3 {
			row, _ := strconv.Atoi(sl1[0])
			col, _ := strconv.Atoi(sl1[1])
			m.Set(row, col, symbolFn(sl1[2]))
		}
	}
}

// 查找图标数量
func (m *Matrix[T]) FindSymbolNum(val T) int {
	var num int
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if m.Get(i, j) == val {
				num++
			}
		}
	}
	return num
}

// 查找图标数量 允许图标转换
func (m *Matrix[T]) FindSymbolNumWithConvert(val T, fn func(val T) T) int {
	var num int
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if fn(m.Get(i, j)) == val {
				num++
			}
		}
	}
	return num
}

// 查找轴val出现次数
func (m *Matrix[T]) FindSymbolByCol(val T, col int) int {
	var num int
	for i := 0; i < m.rows; i++ {
		if m.data[i][col] == val {
			num++
		}
	}
	return num
}

// 参数说明：cleanList清除的单元，返回空的位置信息
// 掉落是指：先清除指定单元元素后，再自然下落，之后空余元素
// 例：如下，先清除为2的元素，然后第一行掉落
// |-\-|--0--|--1--|--2--|--3--|--4--|
// |-0-|  1  |  1  |  1  |  1  |  1  |
// |-1-|  2  |  2  |  2  |  2  |  2  |
// |-2-|  1  |  1  |  1  |  1  |  1  |
// --------------下落结果------------
// |-\-|--0--|--1--|--2--|--3--|--4--|
// |-0-|  0  |  0  |  0  |  0  |  0  |----此行为空出来的元素
// |-1-|  1  |  1  |  1  |  1  |  1  |
// |-2-|  1  |  1  |  1  |  1  |  1  |
func (m *Matrix[T]) Falling(cleanList []*MatrixPosition) []*MatrixPosition {
	for i := 0; i < len(cleanList); i++ {
		m.data[cleanList[i].Row][cleanList[i].Col] = 0
	}
	for i := 0; i < m.cols; i++ {
		var tmp []T
		for j := m.rows - 1; j >= 0; j-- { //从最底部开始
			tmp = append(tmp, m.data[j][i])
			//统一置为0
			m.data[j][i] = 0
		}
		index := m.rows - 1
		for j := 0; j < len(tmp); j++ {
			if tmp[j] == 0 {
				continue
			} else {
				m.data[index][i] = tmp[j]
				index--
			}
		}
	}
	//再遍历一遍获取空元素
	var emptyPosition []*MatrixPosition
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if m.data[i][j] == 0 {
				emptyPosition = append(emptyPosition, &MatrixPosition{Row: i, Col: j})
			}
		}
	}
	return emptyPosition
}

// 叠加图案
// 必须两个完全一样的尺寸
func (m *Matrix[T]) Piling(upMatrix *Matrix[T], blow T) error {
	if len(upMatrix.data) != len(m.data) {
		return ErrInvalidData
	}
	for i := 0; i < len(m.data); i++ {
		if len(upMatrix.data[i]) != len(m.data[i]) {
			return ErrInvalidData
		}
		for j := 0; j < len(m.data[i]); j++ {
			if upMatrix.data[i][j] != 0 {
				m.data[i][j] += upMatrix.data[i][j] * blow // 放大倍数后相加
			} else {
				m.data[i][j] = 0
			}
		}
	}
	return nil
}
