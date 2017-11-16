package main

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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
		hash := GetMD5Hash(input)
		fpgaHash := GetMD5HashFPGA(input)
		c.JSON(http.StatusOK, map[string]string{"cpu": hash, "fpga": fpgaHash, "input": input})
	})
	var port = "80"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	r.Run(":" + port) // listen and serve on 0.0.0.0:80

}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
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
				min-height: 200px;
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
			CPU: &nbsp;<span id="cpu"></span> <br />
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
