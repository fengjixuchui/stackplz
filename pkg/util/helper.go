package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

// 格式化输出相关

const CHUNK_SIZE = 16
const CHUNK_SIZE_HALF = CHUNK_SIZE / 2

const (
	COLORRESET  = "\033[0m"
	COLORRED    = "\033[31m"
	COLORGREEN  = "\033[32m"
	COLORYELLOW = "\033[33m"
	COLORBLUE   = "\033[34m"
	COLORPURPLE = "\033[35m"
	COLORCYAN   = "\033[36m"
	COLORWHITE  = "\033[37m"
)

func dumpByteSlice(b []byte, perfix string) *bytes.Buffer {
	var a [CHUNK_SIZE]byte
	bb := new(bytes.Buffer)
	n := (len(b) + (CHUNK_SIZE - 1)) &^ (CHUNK_SIZE - 1)

	for i := 0; i < n; i++ {

		// 序号列
		if i%CHUNK_SIZE == 0 {
			bb.WriteString(perfix)
			bb.WriteString(fmt.Sprintf("%04d", i))
		}

		// 长度的一半，则输出4个空格
		if i%CHUNK_SIZE_HALF == 0 {
			bb.WriteString("    ")
		} else if i%(CHUNK_SIZE_HALF/2) == 0 {
			bb.WriteString("  ")
		}

		if i < len(b) {
			bb.WriteString(fmt.Sprintf(" %02X", b[i]))
		} else {
			bb.WriteString("  ")
		}

		// 非ASCII 改为 .
		if i >= len(b) {
			a[i%CHUNK_SIZE] = ' '
		} else if b[i] < 32 || b[i] > 126 {
			a[i%CHUNK_SIZE] = '.'
		} else {
			a[i%CHUNK_SIZE] = b[i]
		}

		// 如果到达size长度，则换行
		if i%CHUNK_SIZE == (CHUNK_SIZE - 1) {
			bb.WriteString(fmt.Sprintf("    %s\n", string(a[:])))
		}
	}
	return bb
}

func PrettyByteSlice(buffer []byte) string {
	var out strings.Builder
	for _, b := range buffer {
		if b >= 32 && b <= 126 {
			out.WriteByte(b)
		} else {
			out.WriteString(fmt.Sprintf("\\x%02x", b))
		}
	}
	return out.String()
}

func HexDump(buffer []byte, color string) string {
	b := dumpByteSlice(buffer, color)
	b.WriteString(COLORRESET)
	return b.String()
}

func HexDumpPure(buffer []byte) string {
	b := dumpByteSlice(buffer, "")
	return b.String()
}

func HexDumpGreen(buffer []byte) string {
	b := dumpByteSlice(buffer, COLORGREEN)
	b.WriteString(COLORRESET)
	return b.String()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func IntToBytes(n int) []byte {
	x := uint32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

func UIntToBytes(x uint32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

func RemoveDuplication_map(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}

	return arr[:j]
}

func FindLib(library string, search_paths []string) (string, error) {
	// 尝试在给定的路径中搜索 主要目的是方便用户输入库名即可
	search_paths = RemoveDuplication_map(search_paths)
	// 以 / 开头的认为是完整路径 否则在提供的路径中查找
	if strings.HasPrefix(library, "/") {
		_, err := os.Stat(library)
		if err != nil {
			// 出现异常 提示对应的错误信息
			if os.IsNotExist(err) {
				return library, fmt.Errorf("%s not exists", library)
			}
			return library, err
		}
	} else {
		var full_paths []string
		for _, search_path := range search_paths {
			// 去掉末尾可能存在的 /
			check_path := strings.TrimRight(search_path, "/") + "/" + library
			_, err := os.Stat(check_path)
			if err != nil {
				// 这里在debug模式下打印出来
				continue
			}
			full_paths = append(full_paths, check_path)
		}
		if len(full_paths) == 0 {
			// 没找到
			return library, fmt.Errorf("can not find %s in these paths\n%s", library, strings.Join(search_paths[:], "\n\t"))
		}
		if len(full_paths) > 1 {
			// 在已有的搜索路径下可能存在多个同名的库 提示用户指定全路径
			return library, fmt.Errorf("find %d libs with the same name\n%s", len(full_paths), strings.Join(full_paths[:], "\n\t"))
		}
		// 修正为完整路径
		library = full_paths[0]
	}
	return library, nil
}

func B2STrim(src []byte) string {
	return string(bytes.TrimSpace(bytes.Trim(src, "\x00")))
}

func ParseReg(pid uint32, value uint64) (string, error) {
	info := "UNKNOWN"
	// 直接读取maps信息 计算value在什么地方 用于定位跳转目的地
	filename := fmt.Sprintf("/proc/%d/maps", pid)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return info, fmt.Errorf("Error when opening file:%v", err)
	}
	var (
		seg_start  uint64
		seg_end    uint64
		permission string
		seg_offset uint64
		device     string
		inode      uint64
		seg_path   string
	)
	for _, line := range strings.Split(string(content), "\n") {
		reader := strings.NewReader(line)
		n, err := fmt.Fscanf(reader, "%x-%x %s %x %s %d %s", &seg_start, &seg_end, &permission, &seg_offset, &device, &inode, &seg_path)
		if err == nil && n == 7 {
			if value >= seg_start && value < seg_end {
				offset := seg_offset + (value - seg_start)
				info = fmt.Sprintf("%s + 0x%x", seg_path, offset)
				break
			}
		}
	}
	return info, err
}
