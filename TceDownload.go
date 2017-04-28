package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	archFlag    = flag.String("arch", "x86", "The architecture for which to get extensions.")
	helpFlag    = flag.Bool("help", false, "Shows this help message.")
	kernelFlag  = flag.String("kernel", "4.8.17-tinycore", "The name of the kernel to use for kernel-specific extensions.")
	outFlag     = flag.String("out", "tce/%v/%a", "The directory to which to output files.")
	versionFlag = flag.String("version", "8.x", "The Tiny Core Linux version for which to get extensions.")
)

var baseDir string
var checked = map[string]struct{}{}

func calculateHash(reader io.Reader) (string, error) {
	hash := md5.New()

	_, err := io.Copy(hash, reader)
	if err != nil {
		return "", err
	}

	raw := hash.Sum(nil)
	return hex.EncodeToString(raw), nil
}

func getBaseDir() string {
	return strings.NewReplacer(
		"%a", *archFlag,
		"%v", *versionFlag,
	).Replace(*outFlag)
}

func openFile(fileName string) (io.ReadCloser, error) {
	filePath := filepath.Join(baseDir, fileName)

	fmt.Printf("Checking %v... ", fileName)

	file, err := os.Open(filePath)
	if err == nil {
		info, err := file.Stat()
		if err != nil {
			file.Close()
			return nil, err
		}

		if info.Size() > 0 {
			fmt.Println("Present!")
			return file, nil
		} else {
			fmt.Println("Known absent!")
			return nil, nil
		}
	}

	if !os.IsNotExist(err) {
		fmt.Println("Failed!")
		return nil, err
	}

	fmt.Println("Absent!")
	fmt.Printf("Downloading %v... ", fileName)

	fileUrl := fmt.Sprintf("http://tinycorelinux.net/%v/%v/tcz/%v", *versionFlag, *archFlag, fileName)

	response, err := http.Get(fileUrl)
	if err != nil {
		fmt.Println("Failed!")
		return nil, err
	}
	defer response.Body.Close()

	if (response.StatusCode < 200 || response.StatusCode >= 300) && response.StatusCode != 404 {
		fmt.Println("Failed!")
		return nil, fmt.Errorf("Server returned: %v", response.Status)
	}

	file, err = os.Create(filePath)
	if err != nil {
		fmt.Println("Failed!")
		return nil, err
	}

	if response.StatusCode == 404 {
		fmt.Println("OK!")
		return nil, nil
	}

	_, err = io.Copy(file, response.Body)
	if err != nil {
		file.Close()
		fmt.Println("Failed!")
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		file.Close()
		fmt.Println("Failed!")
		return nil, err
	}

	fmt.Println("OK!")
	return file, nil
}

func getChecksum(name string) (string, error) {
	file, err := openFile(name + ".tcz.md5.txt")
	if err != nil {
		return "", err
	}

	if file == nil {
		return "", nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	if scanner.Scan() {
		return scanner.Text(), nil
	} else {
		return "", nil
	}
}

func getDependencies(name string) ([]string, error) {
	file, err := openFile(name + ".tcz.dep")
	if err != nil {
		return nil, err
	}

	if file == nil {
		return []string{}, nil
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			line = strings.TrimSuffix(line, ".tcz")
			lines = append(lines, line)
		}
	}

	return lines, nil
}

func getExtension(name string) error {
	name = strings.Replace(name, "KERNEL", *kernelFlag, -1)

	if _, ok := checked[name]; ok {
		return nil
	}

	file, err := openFile(name + ".tcz")
	if err != nil {
		return err
	}

	if file == nil {
		return fmt.Errorf("Extension not found: %v", name)
	}
	defer file.Close()

	expectedHash, err := getChecksum(name)
	if err != nil {
		return err
	}

	if expectedHash != "" {
		actualHash, err := calculateHash(file)
		if err != nil {
			return err
		}

		if actualHash != expectedHash {
			return fmt.Errorf("Hash for %v does not match (%v != %v)!", name, actualHash, expectedHash)
		}
	}

	checked[name] = struct{}{}

	dependencies, err := getDependencies(name)
	if err != nil {
		return err
	}

	for _, dependency := range dependencies {
		err = getExtension(dependency)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	if *helpFlag {
		flag.PrintDefaults()
		return
	}

	n := flag.NArg()
	if n == 0 {
		fmt.Printf("USAGE: %v [options] <extension> [extension [...]]\n", os.Args[0])
		fmt.Printf("Invoke %v -help for more information on available options.\n", os.Args[0])
		return
	}

	baseDir = getBaseDir()
	fmt.Printf("Base directory: %v\n", baseDir)

	os.MkdirAll(baseDir, os.ModeDir|0777)

	for _, extension := range flag.Args() {
		err := getExtension(extension)
		if err != nil {
			fmt.Printf("Failed to get %v! %v\n", extension, err.Error())
		} else {
			fmt.Printf("Retrieved %v successfully.\n", extension)
		}
	}
}
