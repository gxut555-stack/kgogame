package common

import (
	"fmt"
	"math/rand"
	"strconv"
)

/**
 * CPF: 巴西个人税务登记号，由11位数字组成，格式：xxx.xxx.xxx.xx 最后两个称为数字验证器 (DV)，由前 9 个数字创建，用于验证整个数字
 * CNPJ：巴西法人税务登记号，由14位数字组成
 * CPF 算法：https://www.oishare.com/tools/cpfdesc.html
 */

type CPF struct {
	Uid int32
	num []int32
}

// 没有uid将，随机生成数字，如果有uid，将根据uid生成
func NewCPF(uid int32) *CPF {
	return &CPF{
		Uid: uid,
	}
}

// 随机生成9位数字
func (c *CPF) genNum() {
	c.num = make([]int32, 0)
	if c.Uid == 0 {
		for i := 0; i < 9; i++ {
			c.num = append(c.num, rand.Int31n(10))
		}
	} else {
		str := fmt.Sprintf("5%08d", c.Uid%100000000)
		for i := 0; i < len(str); i++ {
			n, _ := strconv.Atoi(string(str[i]))
			c.num = append(c.num, int32(n))
		}
	}
}

// 生成验证位
func (c *CPF) genDV() {
	sum := int32(0)
	size := int32(len(c.num))
	for i := int32(0); i < size; i++ {
		sum += (size + 1 - i) * c.num[i]
	}

	rev := 11 - sum%11
	if rev == 10 || rev == 11 {
		c.num = append(c.num, 0)
	} else {
		c.num = append(c.num, rev)
	}
}

func (c *CPF) String() string {
	str := ""
	for _, v := range c.num {
		str += fmt.Sprintf("%d", v)
	}
	return str
}

func (c *CPF) Format() string {
	str := ""
	for i, v := range c.num {
		dot := ""
		if (i+1)%3 == 0 {
			dot = "."
		}
		str += fmt.Sprintf("%d%s", v, dot)
	}
	return str
}

// Gen 生成CPF
// @param bool isFormat 是否格式化
func (c *CPF) Gen(isFormat bool) string {
	c.genNum()
	c.genDV()
	c.genDV()

	if isFormat {
		return c.Format()
	}

	return c.String()
}

// GenCPF 生成CPF
// @param int32 uid 用户uid，如果不传将随机生成
// @param bool isFormat 返回CPF是否格式化
func GenCPF(uid int32, isFormat bool) string {
	return NewCPF(uid).Gen(isFormat)
}
