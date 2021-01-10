package main

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/benbjohnson/hashfs"
	"github.com/gin-gonic/gin"
	"github.com/polaris1119/embed"
)

func main() {
	r := gin.Default()

	r.GET("/assets/*filepath", func(ctx *gin.Context) {
		filename, err := url.PathUnescape(ctx.Param("filepath"))
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		filename = strings.TrimPrefix(filename, "/")

		isHashed := false
		if base, hash := hashfs.ParseName(filename); hash != "" {
			if embed.Fsys.HashName(base) == filename {
				filename = base
				isHashed = true
			}
		}

		f, err := embed.Fsys.Open(filename)
		if os.IsNotExist(err) {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		} else if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer f.Close()

		// Fetch file info. Disallow directories from being displayed.
		fi, err := f.Stat()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		} else if fi.IsDir() {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		contentType := "text/plain"
		// Determine content type based on file extension.
		if ext := path.Ext(filename); ext != "" {
			contentType = mime.TypeByExtension(ext)
		}

		// Cache the file aggressively if the file contains a hash.
		if isHashed {
			ctx.Writer.Header().Set("Cache-Control", `public, max-age=31536000`)
		}

		// Set content length.
		ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

		// Flush header and write content.
		buf := new(bytes.Buffer)
		if ctx.Request.Method != "HEAD" {
			io.Copy(buf, f)
		}
		ctx.Data(http.StatusOK, contentType, buf.Bytes())
	})

	r.LoadHTMLGlob("template/*")
	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"mainjs": embed.Fsys.HashName("static/main.js"),
		})
	})
	r.Run(":8080")
}
