package main

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	SetupFPGA()
	defer CleanupFPGA()

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Visit /md5/{your_string_here} to have an FPGA hash that string")
	})
	r.GET("/md5/:input", func(c *gin.Context) {
		input := c.Param("input")
		hash := GetMD5Hash(input)
		fpgaHash := GetMD5HashFPGA(input)
		c.String(http.StatusOK, "CPU says: %s, FPGA says: %s", hash, fpgaHash)

	})
	r.Run(":80") // listen and serve on 0.0.0.0:80

}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
