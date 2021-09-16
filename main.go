package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const HELP = `
在目录中以正则表达式搜索文件并批量替换为新文件

使用方法：

    1. 无参数交互运行：
    FSR

    2. 在当前目录中搜索：
    FSR [FileNameRegexp]

    3. 在指定目录中搜索：
    FSR [Directory] [FileNameRegexp]

    4. 在指定目录中搜索并替换为指定文件：
    FSR [Directory] [FileNameRegexp] [ReplacementFile]

参数说明：

    [Directory]       ： 搜索目录，支持相对路径，交互模式下默认为当前目录
    [FileNameRegexp]  ： 被搜索的文件名正则表达式
    [Replacement]     ： 被替换的文件，支持相对路径

`

var (
	Directory      string
	FileNameRegexp string
	Replacement    string
)

func help() {
	fmt.Print(HELP)

}

//go:generate goversioninfo
func main() {
	// 命令行参数
	args := os.Args

	switch len(args) {
	case 1:
	case 2:
		if args[1] == "/?" {
			help()
			return
		}
		Directory = "./"
		FileNameRegexp = args[1]
	case 3:
		Directory = args[1]
		FileNameRegexp = args[2]
	case 4:
		Directory = args[1]
		FileNameRegexp = args[2]
		Replacement = args[3]
	default:
		help()
		return
	}

	if validateFromDir(Directory) != "" {
		Directory = ""
	}

	var exit bool

	if Directory == "" {
		if Directory, exit = waitInput("搜索目录", validateFromDir); exit {
			return
		}
	}

	fmt.Printf("搜索目录为：%s\n", Directory)

	if FileNameRegexp == "" {
		if FileNameRegexp, exit = waitInput("搜索文件正则表达式", validateSearchFile); exit {
			return
		}
	}

	fmt.Println("需要被替换的文件如下：")
	loopFind(func(p string, f os.FileInfo) error {
		fmt.Println(p)
		return nil
	})

	if Replacement == "" {
		if FileNameRegexp, exit = waitInput("替换后的文件", validateReplaceFile); exit {
			return
		}
	}

	fmt.Printf("替换后的文件为：%s\n", Replacement)

	var confirm string
	if confirm, exit = waitInput("是否确认要替换（Y-是[默认]，N-否）", validateConfirm); exit || confirm == "N" || confirm == "n" {
		return
	}

	fmt.Println("开始替换")

	loopFind(func(p string, f os.FileInfo) error {
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("%s 删除失败：%v", p, err)
		} else {
			fmt.Printf("%s 已删除！\n", p)
		}

		dirPath := filepath.Dir(p)
		destPath := filepath.Join(dirPath, filepath.Base(Replacement))
		copy(Replacement, destPath)
		fmt.Printf("%s 已复制！\n", destPath)
		return nil
	})

	fmt.Println("替换完成！")
}

func loopFind(deal func(p string, f os.FileInfo) error) {
	serachFileReg, _ := regexp.Compile(FileNameRegexp)
	filepath.Walk(Directory, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if f.IsDir() {
			return nil
		}
		absp, _ := filepath.Abs(p)
		absr, _ := filepath.Abs(Replacement)
		if strings.EqualFold(absp, absr) {
			return nil
		}

		// 匹配目录
		matched := serachFileReg.MatchString(f.Name())
		if !matched {
			return nil
		}

		return deal(p, f)
	})
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	defer destination.Close()
	return io.Copy(destination, source)
}

func waitInput(name string, validate func(input string) string) (string, bool) {
	for {
		fmt.Printf("请输入 %s >", name)
		reader := bufio.NewReader(os.Stdin) //读取输入的内容
		input, _ := reader.ReadString('\n')

		if input == "" {
			continue
		}

		if input == "exit" {
			return "", true
		}
		input = strings.TrimSuffix(input, "\n")
		input = strings.TrimSuffix(input, "\r")
		if errMsg := validate(input); errMsg != "" {
			fmt.Printf("%s 输入错误：%s，请重新输入！\n", name, errMsg)
			continue
		}
		return input, false
	}
}

func validateFromDir(input string) string {
	fi, err := os.Stat(input)
	if err != nil {
		return "不存在或无法读取"
	}

	if !fi.IsDir() {
		return "不是有效的目录"
	}

	return ""
}

func validateSearchFile(input string) string {
	if _, err := regexp.Compile(input); err != nil {
		return "不是有效的正则表达式"
	}

	return ""
}

func validateReplaceFile(input string) string {
	fi, err := os.Stat(input)
	if err != nil {
		return "不存在或无法读取"
	}

	if fi.IsDir() {
		return "不能替换目录"
	}

	return ""
}

func validateConfirm(input string) string {
	if input != "" && input != "Y" && input != "N" && input != "y" && input != "n" {
		return "请输入 Y 或 N "
	}

	return ""
}
