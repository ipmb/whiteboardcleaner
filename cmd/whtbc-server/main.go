package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/yml/whiteboardcleaner"
)

var (
	maxMemory int64 = 1 * 1024 * 1024 // 1MB

	layoutTmpl string = `{{ define "base" }}<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8"/>
		<title>{{ template "title" .}}</title>
		<link rel="stylesheet" href="/assets/css/main.css" />
	</head>
	<body>
	{{ template "content" . }}
	<div id="upload"></div>
	<script src="/assets/js/bundle.js"></script>
	</body>
</html>
{{ end }}
`

	resultTmpl string = `{{ define "title" }}Whiteboord cleaner | result{{ end }}
{{ define "content" }}{{ range . }}<div><img src="{{ . }}"/></div>{{ end }}{{ end}}
`

	indexTmpl string = `{{ define "title" }}Whiteboard cleaner{{ end }}
{{ define "content" }}
	<form action="/upload/" method="POST" enctype="multipart/form-data">
		<fieldset>
		<legend>Edge detection</legend>
		{{ if .Errors.EdgeDetectionKernelSize }}<div class="error">{{ .Errors.EdgeDetectionKernelSize }}</div>{{ end }}
		<label for="EdgeDetectionKernelSize">EdgeDetectionKernelSize</label>
		<input name="EdgeDetectionKernelSize" type="text" value="{{ .Opts.EdgeDetectionKernelSize }}"></input>

		{{ if .Errors.ConvolutionMultiplicator }}<div class="error">{{ .Errors.ConvolutionMultiplicator }}</div>{{ end }}
		<label for="ConvolutionMultiplicator">ConvolutionMultiplicator</label>
		<input name="ConvolutionMultiplicator" type="text" value="{{ .Opts.ConvolutionMultiplicator }}"></input>
		</fieldset>

		<fieldset>
		<legend>cleanup the image to get a white backgound</legend>

		{{ if .Errors.GaussianBlurSigma }}<div class="error">{{ .Errors.GaussianBlurSigma }}</div>{{ end }}
		<label for="GaussianBlurSigma">GaussianBlurSigma</label>
		<input name="GaussianBlurSigma" type="text" value="{{ .Opts.GaussianBlurSigma }}"></input>

		{{ if .Errors.SigmoidMidpoint }}<div class="error">{{ .Errors.SigmoidMidpoint }}</div>{{ end }}
		<label for="SigmoidMidpoint">SigmoidMidpoint</label>
		<input name="SigmoidMidpoint" type="text" value="{{ .Opts.SigmoidMidpoint }}"></input>

		{{ if .Errors.MedianKsize }}<div class="error">{{ .Errors.MedianKsize }}</div>{{ end }}
		<label for="MedianKsize">MedianKsize</label>
		<input name="MedianKsize" type="text" value="{{ .Opts.MedianKsize }}"></input>
		</fieldset>

		<fieldset>
		<legend>Image</legend>
		{{ if .Errors.file }}<div class="error">{{ .Errors.file }}</div>{{ end }}
		<label for="file">File:</label>
		<input name="file" type="file" id="fileField"></input>
		</fieldset>

		<input type="submit"></input>
	</form>
{{ end }}
`
)

type appContext struct {
	TmpDir                          string
	PrefixTmpDir                    string
	UploadURL, ResultURL, StaticURL, AssetsURL string
	Templates                       map[string]*template.Template
}

func uploadHandler(ctx *appContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		filterOpts := whiteboardcleaner.NewOptions()
		errors := filterOpts.ValidAndUpdate(r.MultipartForm.Value)
		if len(errors) > 0 {
			tmpl := ctx.Templates["index"]
			tmpl.ExecuteTemplate(
				w,
				"base",
				struct {
					Opts   *whiteboardcleaner.Options
					Errors map[string]string
				}{Opts: filterOpts, Errors: errors})
			return
		}

		dirPath, err := ioutil.TempDir(ctx.TmpDir, ctx.PrefixTmpDir)
		//_, dirName := filepath.Split(dirPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		for _, fileHeaders := range r.MultipartForm.File {
			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				tf, err := ioutil.TempFile(dirPath, fmt.Sprintf("%s_", fileHeader.Filename))

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				io.Copy(tf, file)
				// rewind the file to the  begining
				tf.Seek(0, 0)
				// Decode the image
				img, err := jpeg.Decode(tf)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				g := whiteboardcleaner.NewFilter(filterOpts)
				dst := image.NewRGBA(g.Bounds(img.Bounds()))
				g.Draw(dst, img)
				// Create the dstTemporaryFile
				dstTemporaryFile, err := ioutil.TempFile(dirPath, fmt.Sprintf("cleaned_%s_", fileHeader.Filename))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				jpeg.Encode(dstTemporaryFile, dst, &jpeg.Options{Quality: 99})
				imagePath, err := filepath.Rel(os.TempDir(), dstTemporaryFile.Name())
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				w.Write([]byte(filepath.Join(ctx.StaticURL, imagePath)))
			}
		}
	}
}

func resultHandler(ctx *appContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		dirName, err := filepath.Rel(ctx.ResultURL, path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		files, err := filepath.Glob(filepath.Join(ctx.TmpDir, dirName, "*"))
		for i, file := range files {
			rel, err := filepath.Rel(os.TempDir(), file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			files[i] = filepath.Join(ctx.StaticURL, rel)
		}
		tmpl := ctx.Templates["result"]
		tmpl.ExecuteTemplate(w, "base", files)
	}
}

func indexHandler(ctx *appContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		filterOpts := whiteboardcleaner.NewOptions()
		errors := make(map[string]string)
		tmpl := ctx.Templates["index"]
		tmpl.ExecuteTemplate(
			w,
			"base",
			struct {
				Opts   *whiteboardcleaner.Options
				Errors map[string]string
			}{Opts: filterOpts, Errors: errors})
	}
}

type loggedResponseWriter struct {
	http.ResponseWriter
	status int
}

func (l *loggedResponseWriter) WriteHeader(status int) {
	l.status = status
	l.ResponseWriter.WriteHeader(status)
}

func wrap(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggedResponseWriter{ResponseWriter: w, status: http.StatusOK}
		l := log.New(os.Stdout, "[whiteboardcleaner] ", 0)
		h.ServeHTTP(lw, r)
		l.Printf("%s %s %d %s\n", r.Method, r.URL, lw.status, time.Since(start))
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	addr := flag.String("addr", ":8080", "path to the source image")
	flag.Parse()
	asset_path := flag.String("asset_path", "./src/github.com/yml/whiteboardcleaner/assets", "path to static assets")
	tmpls := make(map[string]*template.Template)
	layout := template.Must(template.New("Layout").Parse(layoutTmpl))
	tmpl := template.Must(layout.Clone())
	tmpls["index"] = template.Must(tmpl.New("index").Parse(indexTmpl))
	tmpl = template.Must(layout.Clone())
	tmpls["result"] = template.Must(tmpl.New("result").Parse(resultTmpl))

	ctx := &appContext{
		TmpDir:       os.TempDir(),
		PrefixTmpDir: "whiteboardcleaner_",
		UploadURL:    "/upload/",
		ResultURL:    "/cleaned/",
		StaticURL:    "/static/",
		AssetsURL:    "/assets/",
		Templates:    tmpls,
	}

	fmt.Println("Starting whiteboard cleaner server listening on addr", *addr)

	mux := http.NewServeMux()
	mux.HandleFunc(ctx.UploadURL, wrap(uploadHandler(ctx)))
	mux.HandleFunc(ctx.ResultURL, wrap(resultHandler(ctx)))
	mux.Handle(ctx.StaticURL,
		http.StripPrefix(ctx.StaticURL, http.FileServer(http.Dir(os.TempDir()))))
	mux.Handle(ctx.AssetsURL,
		http.StripPrefix(ctx.AssetsURL, http.FileServer(http.Dir(*asset_path))))
	mux.HandleFunc("/", wrap(indexHandler(ctx)))
	http.ListenAndServe(*addr, mux)
}
