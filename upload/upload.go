package upload

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Progress struct {
	TotalSize int64
	BytesRead int64
}

const UPLOAD_PATH = "./uploads"

func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}

	result := float64(pr.BytesRead) / float64(pr.TotalSize)
	percen := result * 100
	fmt.Println("progress = ")
	fmt.Println(percen)

	//fmt.Printf("File upload in progress: %d\n", pr.BytesRead)
}

func Upload(r *http.Request) (string, error) {
	start := time.Now()

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return "", err
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	// filetype := http.DetectContentType(buff)
	// if filetype != "text/csv" {
	// 	fmt.Println("file type is = ")
	// 	fmt.Println(filetype)
	// 	fmt.Println("file name is = ")
	// 	fmt.Println(fileHeader.Filename)
	// 	fmt.Println("file content-type is = ")
	// 	fmt.Println(fileHeader.Header.Get("Content-Type"))
	// 	fmt.Println(fileHeader.Size)
	// 	http.Error(w, "The provided file format is not allowed. Please upload a CSV file", http.StatusBadRequest)
	// 	return
	// }

	filetype := fileHeader.Header.Get("Content-Type")
	if filetype != "text/csv" {
		fmt.Println("file type is = ")
		fmt.Println(filetype)
		//http.Error(w, "The provided file format is not allowed. Please upload a CSV file", http.StatusBadRequest)
		err = errors.New("the provided file format is not allowed. please upload a csv file")
		return "", err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	// Create the uploads folder if it doesn't
	// already exist
	err = os.MkdirAll(UPLOAD_PATH, os.ModePerm)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	// Create a new file in the uploads directory
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
	dst, err := os.Create(UPLOAD_PATH + "/" + fileName)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	defer dst.Close()

	pr := &Progress{
		TotalSize: fileHeader.Size,
	}

	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, io.TeeReader(file, pr))
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	usage := time.Since(start)
	fmt.Println("usage", usage)

	return fileName, nil
}
