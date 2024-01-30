package domain

type OCRMethod string

const (
	NewOCRMethod       OCRMethod = "new_ocr"
	TesseractOCRMethod           = "tesseract"
	IOCRMethod                   = "iOCR"
	CloudNumberKey               = "CLOUD_NUMBER"
)
