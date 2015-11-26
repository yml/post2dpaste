package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

var (
	dpasteUrl string = "https://dpaste.de/api/"
	lexer     string
	filename  string
)

func init() {
	flag.StringVar(&lexer, "lexer", "default", "lexer options are: python, go, c, mysql, ...")
}

func main() {
	flag.Parse()
	var bufInput bytes.Buffer

	var writers []io.Writer
	writers = append(writers, os.Stdout)
	writers = append(writers, &bufInput)
	mWriter := io.MultiWriter(writers...)

	if len(os.Args) == 1 {
		mReader := io.MultiReader(os.Stdin, os.Stderr)
		if _, err := io.Copy(mWriter, mReader); err != nil {
			log.Fatal("Error while copying from stdin to stdout", err)
		}
	} else {
		for _, filename = range os.Args[1:] {
			fh, err := os.Open(filename)
			if err != nil {
				log.Fatal("Error while opening:", filename, err)
			}
			_, err = io.Copy(mWriter, fh)
			if err != nil {
				log.Fatal("Error while copying:", filename, err)
			}
		}
	}

	if lexer == "default" && filename == "" {
		lexer = "GO"
	} else if lexer == "default" && filename != "" {
		lexer = ""
	}

	u, err := url.ParseRequestURI(dpasteUrl)
	if err != nil {
		log.Fatal("Error while parsing dpasteUrl", err)
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add field content
	fw, err := w.CreateFormField("content")
	if err != nil {
		log.Fatal("Error while creating `content` field", err)
	}
	if _, err := fw.Write(bufInput.Bytes()); err != nil {
		log.Fatal("Error while writing to `content` field", err)
	}
	fw, err = w.CreateFormField("lexer")
	if err != nil {
		log.Fatal("Error while creating `lexer` field", err)
	}
	if _, err := fw.Write([]byte(lexer)); err != nil {
		log.Fatal("Error while writing to `lexer` field", err)
	}
	if filename != "" {
		fw, err = w.CreateFormField("filename")
		if err != nil {
			log.Fatal("Error while creating `filename` field", err)
		}
		if _, err := fw.Write([]byte(filename)); err != nil {
			log.Fatal("Error while writing to `filename` field", err)
		}
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", u.String(), &b)
	if err != nil {
		log.Fatalf("Error while building the request to dpaste:", err)
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error while posting to dpaste:", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error while reading the response Body:", err)
	}
	fmt.Println("\n\ndpasted :", string(body)[1:len(string(body))-1])
}
