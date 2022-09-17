package main

import (
	"fmt"
	"math"
	"strconv"
)

func MsgFilling(plainText string) []string { // 消息填充
	final_data := make([]string, 0)
	plainData := []byte(plainText)
	data := make([]string, 0)
	for _, v := range plainData {
		tempData := strconv.FormatInt(int64(v), 2)
		for len(tempData) < 8 {
			tempData = "0" + tempData
		}
		for _, v := range tempData {
			data = append(data, string(v))
		}
	}
	lenth := len(data)
	lenth2bin := strconv.FormatInt(int64(lenth), 2)
	// +1
	data = append(data, "1")
	k := 0
	for ; (lenth+1+k)%512 != 448; k++ {
		data = append(data, "0")
	}
	for len(lenth2bin) < 64 {
		lenth2bin = "0" + lenth2bin
	}
	// 循环插入
	for _, v := range lenth2bin {
		data = append(data, string(v))
	}
	// 对data进行4位处理
	for i := 0; i*8 < len(data); i++ {
		res := ""
		for _, v := range data[i*8 : i*8+8] {
			res += v
		}
		num, _ := strconv.ParseInt(res, 2, 64)
		final_num := strconv.FormatInt(int64(num), 16)
		// 填充16进制
		for len(final_num) < 2 {
			final_num = "0" + final_num
		}
		final_data = append(final_data, final_num)
	}
	return final_data
}

// 消息分组
func OrgMsg(originData []string) [][132]string {
	var b = [...][16]string{{}}
	var data = [...][132]string{{}}
	for i := 0; i*64 < len(originData); i++ {
		for j := 0; j*4 < 64; j++ {
			b[i][j] = (originData[j*4] + originData[j*4+1] + originData[j*4+2] + originData[j*4+3])
		}
	}
	for l := 0; l < (len(b)); l++ {
		data[l] = msgExpand(b[l])
	}
	return data[:]
}

// 将int的数据恢复到str16位
func reStr(x int64) string {
	data := strconv.FormatInt(x, 2)
	res := ""
	for len(data) < 32 {
		data = "0" + data
	}
	for i := 0; i*4 < len(data); i++ {
		tempData := data[i*4:i*4+1] + data[i*4+1:i*4+2] + data[i*4+2:i*4+3] + data[i*4+3:i*4+4]
		tem, _ := strconv.ParseInt(tempData, 2, 64)
		res += strconv.FormatInt(tem, 16)
	}
	return res
}

// P0函数
func P0(X int64) (data string) {
	data = reStr(X ^ dataShift(reStr(X), 9) ^ dataShift(reStr(X), 17))
	return
}

// P1函数
func P1(x int64) int64 {
	strX := reStr(x)
	return x ^ dataShift(strX, 15) ^ dataShift(strX, 23)
}

// 消息扩展
func msgExpand(b [16]string) [132]string {
	w := [132]string{}
	for i := 0; i < 16; i++ {
		w[i] = b[i]
	}
	for j := 16; j < 68; j++ {
		tempW := ""
		tempW = reStr(P1(dataShift(w[j-16], 0)^dataShift(w[j-9], 0)^dataShift(w[j-3], 15)) ^ dataShift(w[j-13], 7) ^ dataShift(w[j-6], 0))
		w[j] = tempW
	}
	for k := 68; k < 132; k++ {
		w[k] = reStr(dataShift(w[k-68], 0) ^ dataShift(w[k-64], 0))
	}
	return w
}

// mod2^32加法运算
func mod32(x int64, y int64) int64 {
	return (x + y) % int64(math.Pow(2, 32))
}

// T函数
func T(j int) (data int64) {
	if j < 16 {
		data = dataShift("79cc4519", j)
	} else {
		j = j % 32
		data = dataShift("7a879d8a", j)
	}
	return
}

// FF函数
func FF(A string, B string, C string, j int) (data int64) {
	X := dataShift(A, 0)
	Y := dataShift(B, 0)
	Z := dataShift(C, 0)
	if j < 16 {
		data = X ^ Y ^ Z
	} else {
		data = (X & Y) | (X & Z) | (Y & Z)
	}
	return
}

func GG(A string, B string, C string, j int) (data int64) {
	X := dataShift(A, 0)
	Y := dataShift(B, 0)
	Z := dataShift(C, 0)
	if j < 16 {
		data = X ^ Y ^ Z
	} else {
		data = (X & Y) | (^X & Z)
	}
	return
}

// msgZip 数据压缩 64轮
func msgZip(w [132]string, v []string) []string {
	data := make([]string, 0)
	var A, B, C, D, E, F, G, H = v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7]
	for i := 0; i < 64; i++ {
		SS1 := dataShift(reStr(mod32(mod32(dataShift(A, 12), dataShift(E, 0)), T(i))), 7)
		SS2 := SS1 ^ dataShift(A, 12)
		TT1 := mod32(mod32(mod32(FF(A, B, C, i), dataShift(D, 0)), SS2), dataShift(w[i+68], 0))
		TT2 := mod32(GG(E, F, G, i), mod32(mod32(SS1, dataShift(w[i], 0)), dataShift(H, 0)))
		D = C
		C = reStr(dataShift(B, 9))
		B = A
		A = reStr(TT1)
		H = G
		G = reStr(dataShift(F, 19))
		F = E
		E = P0(TT2)
	}
	var dataList = []string{A, B, C, D, E, F, G, H}
	// fmt.Println(dataList)
	for i := 0; i < len(v); i++ {
		data = append(data, reStr(dataShift(v[i], 0)^dataShift(dataList[i], 0)))
	}
	return data
}

// dataShift 数据移位
func dataShift(data string, l int) int64 {
	tempData := ""
	for _, v := range data {
		num, _ := strconv.ParseInt(string(v), 16, 64)
		num4 := strconv.FormatInt(num, 2)
		for len(num4) < 4 {
			num4 = "0" + num4
		}
		tempData += num4
	}
	for len(tempData) < 32 {
		tempData = "0" + tempData
	}
	if l != 0 {
		tempData = tempData[l:] + tempData[:l]
	}
	tempRes := ""
	for i := 0; i*4 < len(tempData); i++ {
		tData := tempData[i*4:i*4+1] + tempData[i*4+1:i*4+2] + tempData[i*4+2:i*4+3] + tempData[i*4+3:i*4+4]
		tem, _ := strconv.ParseInt(tData, 2, 64)
		tempRes += strconv.FormatInt(tem, 16)
	}
	res, _ := strconv.ParseInt(tempRes, 16, 64)
	return res
}

func main() {
	originData := MsgFilling("abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd")
	var originV = []string{"7380166f", "4914b2b9", "172442d7", "da8a0600", "a96f30bc", "163138aa", "e38dee4d", "b0fb0e4e"}
	V := make([][]string, 0, 0)
	V = append(V, originV)

	for l := 0; l*64 < len(originData); l++ {
		data := originData[l*64 : (l+1)*64]
		w := OrgMsg(data)[0]
		V = append(V, msgZip(w, V[l]))
	}
	res := V[len(V)-1:]
	final_data := ""
	for _, v := range res[0] {
		final_data += v
	}
	fmt.Println(final_data)
}
