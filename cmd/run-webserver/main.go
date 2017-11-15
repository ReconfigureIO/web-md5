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
	<html>
	<head>
		<title>Web MD5 Example</title>
		<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
	</head>
	<body>
		<h3>Web MD5 Example</h3>
		<div>
			Input:
			<br />
			<textarea id="input" name="input" style="border: solid 1px #555;"></textarea>
		</div>
		<button id="compute">Compute MD5</button>
		<div style="font-family: Monospace; display: none; padding-top: 10px;" id="resultDiv">
			CPU: &nbsp;<span id="cpu"></span> <br />
			FPGA: <span id="fpga"></span>
		</div>
		<script src="https://code.jquery.com/jquery-3.2.1.min.js"></script>
		<script>
			function compute(){
				$('#compute').attr('disabled', 'disabled');
				$('#compute').text("Computing...");
				$('#resultDiv').hide();

				$.post('/md5', {input: $('#input').val()}, function(resp){
					showResult(resp);
				});
			}

			function showResult(result){
				$('#compute').removeAttr('disabled');
				$('#compute').text("Compute MD5");
				$('#cpu').html(result.cpu);
				$('#fpga').html(result.fpga);
				$('#resultDiv').show();
			}

			$('#compute').click(compute);
		</script>
	</body>
	</html>
`
