package test

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"testing"
)

// 测试文件分片
// 分片大小为100MB
var chunkSize = 100 * 1024 * 1024 //100MB

func TestFileChunk(t *testing.T) {
	//获取文件状态
	fileInfo, err := os.Stat("")

	if err != nil {
		return
	}

	//计算文件分片个数
	//向上取整
	chunkNum := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))

	//开辟字节切片 大小为分片大小
	b := make([]byte, chunkSize)

	//读取文件
	file, err := os.OpenFile("", os.O_RDONLY, 0777)
	//循环将文件进行分片
	for i := 0; i < chunkNum; i++ {
		//指定读取文件的起始位置
		file.Seek(int64(i*chunkSize), 0)
		if err != nil {
			return
		}
		defer file.Close()

		if fileInfo.Size()-int64(chunkSize*i) < int64(chunkSize) {
			b = make([]byte, fileInfo.Size()-int64(chunkSize*i))
		}
		file.Read(b)
		//制作文件分片
		chunkFile, err := os.OpenFile("./"+strconv.Itoa(i)+".chunk", os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			return
		}
		defer chunkFile.Close()
		_, err = chunkFile.Write(b)
		if err != nil {
			return
		}
		fmt.Println("文件分片" + strconv.Itoa(i) + "已经写完")

	}

	//关闭资源
}

//测试文件分片合并

func TestFileChunkMerge(t *testing.T) {
	file, err := os.OpenFile("", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return
	}
	defer file.Close()
	//获取文件状态
	fileInfo, err := os.Stat("")

	if err != nil {
		return
	}

	chunkNum := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))
	for i := 0; i < chunkNum; i++ {
		//向文件中写
		chunkFile, err := os.OpenFile("./"+strconv.Itoa(i)+".chunk", os.O_RDONLY, 0777)
		if err != nil {
			return
		}
		defer chunkFile.Close()
		b, err := ioutil.ReadAll(chunkFile)
		if err != nil {
			return
		}
		file.Write(b)
	}

}

//测试文件一致性

func TestFileSame(t *testing.T) {
	//校验MD5 字节和是否一致
	file1, err := os.OpenFile("", os.O_RDONLY, 0777)
	if err != nil {
		return
	}
	defer file1.Close()
	b1, err := ioutil.ReadAll(file1)
	if err != nil {
		return
	}
	f1Md5Sum := md5.Sum(b1)
	file2, err := os.OpenFile("", os.O_RDONLY, 0777)
	if err != nil {
		return
	}
	defer file2.Close()
	b2, err := ioutil.ReadAll(file2)
	if err != nil {
		return
	}
	f2Md5Sum := md5.Sum(b2)
	f1Str := fmt.Sprintf("%x", f1Md5Sum)
	f2Str := fmt.Sprintf("%x", f2Md5Sum)
	fmt.Println(f1Str == f2Str)
}
