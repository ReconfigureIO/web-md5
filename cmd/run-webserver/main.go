package main

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/ReconfigureIO/sdaccel/xcl"
	"log"
	"net/http"
	"os"

	"github.com/ReconfigureIO/crypto/md5/host"
	"github.com/gin-gonic/gin"
)

var (
	world   xcl.World
	program *xcl.Program
)

func main() {
	SetupFPGA()
	defer CleanupFPGA()

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-type", "text/html")
		c.String(200, webTemplate)
	})
	r.POST("/md5", func(c *gin.Context) {
		input := c.PostForm("input")
		fpgaHash := GetMD5HashFPGA(input)
		c.JSON(http.StatusOK, map[string]string{"fpga": fpgaHash, "input": input})
	})
	var port = "80"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	r.Run(":" + port) // listen and serve on 0.0.0.0:80

}

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

const webTemplate = `
	<!DOCTYPE html>
	<html>
	<head>
		<title>MD5 Generator</title>
		<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
		<meta charset="utf-8" />
		<link rel="stylesheet" type="text/css" href="//fonts.googleapis.com/css?family=Open+Sans" />
		<style>
			body {
				font-family: 'open sans', sans-serif;
			}
			#compute {
				background-color: #37ca97;
				color: #fff;
				padding: 15px;
				border: none;
				font-size: 1.2em;
				margin: 50px;
				width: 200px;
			}
			textarea {
				background-color: #eee;
				width: 200px;
				border: solid 1px #eee;
				border-radius: 10px;
				padding: 15px;
				font-family: 'Open Sans', sans-serif;
			}
			#input {
				min-height: 100px;
			}
			#output {
				min-height: 100px;
			}
		</style>
	</head>
	<body>
		<div><img src="https://reconfigure.io/assets/dist/img/global/logo.svg" /></div>
		<h2 style="color: #e9ab70;">FPGA based MD5 Generator</h2>
		<div>
			<p>Enter your text in the box on the left and click GENERATE to see the MD5 hash.</p>
			<table>
				<tbody>
					<tr valign="top">
					<td> <textarea id="input" name="input"></textarea> </td>
					<td> <button id="compute" normal-label="GENERATE" busy-label="GENERATING..." >GENERATE</button> </td>
					<td> <textarea id="output"></textarea> </td>
					</tr>
				</tbody>
			</table>
		</div>

		<div style="font-family: Monospace; display: none; padding-top: 10px;" id="resultDiv">
			FPGA: <span id="fpga"></span>
		</div>
		<script src="https://code.jquery.com/jquery-3.2.1.min.js"></script>
		<script>
			function compute(){
				$('#compute').attr('disabled', 'disabled');
				$('#compute').text($('#compute').attr('busy-label'));
				$('#output').val("");

				$.post('/md5', {input: $('#input').val()}, function(resp){
					showResult(resp);
				});
			}

			function showResult(result){
				$('#compute').removeAttr('disabled');
				$('#compute').text($('#compute').attr('normal-label'));
				$('#output').val(result.fpga);
				$('#input').focus();
			}

			$('#compute').click(compute);
		</script>
	</body>
	</html>
`
