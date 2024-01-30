package components

import (
	"ImageProcessor/domain"
	"log"
	"runtime"
)

func GetLine() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}

func GetFile() string {
	_, file, _, _ := runtime.Caller(1)
	return file
}

func LogPrintInfo(title string, line int, file string) {
	log.Printf("--- title: %s, line: %d, file: %s\n", title, line, file)
}

func LogPrintError(title string, err string, line int, file string) {
	log.Printf("*** title: %s, error: %s, line: %d, file: %s\n", title, err, line, file)
}

func DefineMethod(method string) domain.OCRMethod {
	switch method {
	case "new_ocr":
		return domain.NewOCRMethod
	case "iOCR":
		return domain.IOCRMethod
	case "tesseract":
		return domain.TesseractOCRMethod
	}
	return domain.TesseractOCRMethod
}
