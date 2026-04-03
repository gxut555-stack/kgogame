package common

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

func Compress(w io.Writer, data []byte) error {
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gw.Close()
	_, err = gw.Write(data)
	return err
}

func DeCompress(w io.Writer, data []byte) error {
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer gr.Close()
	data, err = io.ReadAll(gr)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// gzip压缩数据并返回base64 encode 后的数据
func JsonGZipCompressAndBase64Encode(val interface{}) (string, error) {
	encodeData, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	//defer gz.Close()
	_, err = gz.Write(encodeData)
	if err != nil {
		gz.Close()
		return "", err
	}
	if err = gz.Close(); err != nil {
		return "", err
	}
	gzipBytes := buf.Bytes()
	base64Bytes := base64.StdEncoding.EncodeToString(gzipBytes)
	buf.Reset()     //fast leak
	gzipBytes = nil //fast leak
	return base64Bytes, nil
}

const (
	MAGIC_HEAD_CODE = 0x55
	MAGIC_TAIL_CODE = 0xAA
)

func CompressWithMagic(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	var rawDataBuf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&rawDataBuf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewWriterLevel(&rawDataBuf, %d)|err: %s", gzip.BestCompression, err.Error())
	}
	_, err = zw.Write(raw)
	if err != nil {
		return nil, fmt.Errorf("zw.Write(req.RawData) failed|err: %s", err.Error())
	}
	_ = zw.Close()
	res := make([]byte, 0, rawDataBuf.Len()+4)
	res = append(res, MAGIC_HEAD_CODE, MAGIC_HEAD_CODE)
	res = append(res, rawDataBuf.Bytes()...)
	res = append(res, MAGIC_TAIL_CODE, MAGIC_TAIL_CODE)
	return res, nil
}

func DecompressWithMagic(data []byte) ([]byte, error) {
	length := len(data)
	if length <= 4 {
		return nil, nil
	}
	if !(data[0] == MAGIC_HEAD_CODE && data[1] == MAGIC_HEAD_CODE && data[length-2] == MAGIC_TAIL_CODE && data[length-1] == MAGIC_TAIL_CODE) {
		return nil, nil
	}

	body := data[2 : length-2]

	var buf bytes.Buffer
	_, err := buf.Write(body)
	if err != nil {
		return nil, fmt.Errorf("buf.Read()|error: %s", err.Error())
	}
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		return nil, nil
	}
	if err = zr.Close(); err != nil {
		return nil, fmt.Errorf("zr.Close()|error:%s", err)
	}
	var rBuf bytes.Buffer
	if _, err = io.Copy(&rBuf, zr); err != nil {
		return nil, fmt.Errorf("io.Copy|error:%s", err)
	}
	return rBuf.Bytes(), nil
}

func GzipCompress(data []byte, level int) ([]byte, error) {
	if level == 0 {
		level = gzip.DefaultCompression
	}
	if len(data) == 0 || data == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	gw, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}
	if _, err = gw.Write(data); err != nil {
		_ = gw.Close()
		return nil, err
	}
	if err = gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GzipDeCompress(data []byte) ([]byte, error) {
	//尝试是否magic压缩包
	length := len(data)
	if length >= 4 && data[0] == MAGIC_HEAD_CODE && data[1] == MAGIC_HEAD_CODE &&
		data[length-2] == MAGIC_TAIL_CODE && data[length-1] == MAGIC_TAIL_CODE { //是magic包，则只用magic包解析
		return DecompressWithMagic(data)
	}
	var buf bytes.Buffer
	_, err := buf.Write(data)
	if err != nil {
		return nil, fmt.Errorf("buf.Read()|error: %s", err.Error())
	}
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		return nil, nil
	}
	if err = zr.Close(); err != nil {
		return nil, fmt.Errorf("zr.Close()|error:%s", err)
	}
	var rBuf bytes.Buffer
	if _, err = io.Copy(&rBuf, zr); err != nil {
		return nil, fmt.Errorf("io.Copy|error:%s", err)
	}
	return rBuf.Bytes(), nil
}
