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
	"text/template"

	"github.com/benbjohnson/hashfs"
	"github.com/labstack/echo/v4"
	"github.com/polaris1119/embed"
)

func main() {
	e := echo.New()

	e.GET("/assets/*", func(ctx echo.Context) error {
		filename, err := url.PathUnescape(ctx.Param("*"))
		if err != nil {
			return err
		}

		isHashed := false
		if base, hash := hashfs.ParseName(filename); hash != "" {
			if embed.Fsys.HashName(base) == filename {
				filename = base
				isHashed = true
			}
		}

		f, err := embed.Fsys.Open(filename)
		if os.IsNotExist(err) {
			return echo.ErrNotFound
		} else if err != nil {
			return echo.ErrInternalServerError
		}
		defer f.Close()

		// Fetch file info. Disallow directories from being displayed.
		fi, err := f.Stat()
		if err != nil {
			return echo.ErrInternalServerError
		} else if fi.IsDir() {
			return echo.ErrForbidden
		}

		contentType := "text/plain"
		// Determine content type based on file extension.
		if ext := path.Ext(filename); ext != "" {
			contentType = mime.TypeByExtension(ext)
		}

		// Cache the file aggressively if the file contains a hash.
		if isHashed {
			ctx.Response().Header().Set("Cache-Control", `public, max-age=31536000`)
		}

		// Set content length.
		ctx.Response().Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

		// Flush header and write content.
		buf := new(bytes.Buffer)
		if ctx.Request().Method != "HEAD" {
			io.Copy(buf, f)
		}
		return ctx.Blob(http.StatusOK, contentType, buf.Bytes())
	})

	e.GET("/", func(ctx echo.Context) error {
		tpl, err := template.New("index.html").ParseFiles("template/index.html")
		if err != nil {
			return err
		}

		var buf = new(bytes.Buffer)
		err = tpl.Execute(buf, map[string]interface{}{
			"mainjs": embed.Fsys.HashName("static/main.js"),
		})
		if err != nil {
			return err
		}
		return ctx.HTML(http.StatusOK, buf.String())
	})

	e.Logger.Fatal(e.Start(":8080"))
}
