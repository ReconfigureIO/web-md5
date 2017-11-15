//+build !local

package main

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"xcl"

	"github.com/ReconfigureIO/crypto/md5/host"
)

var (
	world   xcl.World
	program xcl.Program
)

func SetupFPGA() {
	world = xcl.NewWorld()
	program = world.Import("kernel_test")

}

func GetMD5HashFPGA(text string) string {
	krnl := program.GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	msg := host.Pad([]byte(text))
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
