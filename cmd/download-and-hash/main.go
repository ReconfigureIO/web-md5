package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ReconfigureIO/crypto/md5/host"
	"github.com/ReconfigureIO/sdaccel/xcl"
)

func main() {

	SetupFPGA()
	defer CleanupFPGA()

	fileURL := os.Args[1]

	fmt.Printf("Downloading from URL: %v \n", fileURL)
	data, err := DownloadData(fileURL)
	if err != nil {
		panic(err)
	}

	fmt.Println("Download complete. Computing hash")
	hash := GetMD5HashFPGA(data)
	fmt.Printf("Hashing complete. MD5 of %v is %v \n", fileURL, hash)
}

var (
	world   xcl.World
	program *xcl.Program
)

func SetupFPGA() {
	world = xcl.NewWorld()
	program = world.Import("kernel_test")

}

func GetMD5HashFPGA(data []byte) string {
	krnl := program.GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	msg := host.Pad([]byte(data))
	msgSize := binary.Size(msg)

	inputBuff := world.Malloc(xcl.ReadOnly, uint(msgSize))
	defer inputBuff.Free()

	outputBuff := world.Malloc(xcl.ReadOnly, 16)
	defer outputBuff.Free()

	binary.Write(inputBuff.Writer(), binary.LittleEndian, msg)
	numBlocks := uint32(msgSize / 64)

	krnl.SetArg(0, numBlocks)
	krnl.SetMemoryArg(1, inputBuff)
	krnl.SetMemoryArg(2, outputBuff)

	krnl.Run(1, 1, 1)

	ret := make([]byte, 16)
	err := binary.Read(outputBuff.Reader(), binary.LittleEndian, ret)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	return hex.EncodeToString(ret)
}

func CleanupFPGA() {
	program.Release()
	world.Release()
}

// DownloadData will download the body of the HTTP response to memory
func DownloadData(url string) ([]byte, error) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	return data, err
}
