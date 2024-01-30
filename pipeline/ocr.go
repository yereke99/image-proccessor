package pipeline

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"

	//"github.com/DevClusterRu/gosseract/v3"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os/exec"
	"regexp"
	"strings"

	//"github.com/DevClusterRu/gosseract/v3"
	//_ "github.com/DevClusterRu/gosseract/v3"
	"github.com/disintegration/imaging"
	"github.com/gogf/gf/v2/text/gstr"
)

type Tesseract struct {
	//Client *gosseract.Client
}

var (
	ErrRecognition = errors.New("error recognition")
)

const (
	BasePath     = "/Users/badmin/ImageProcessor"
	TextStopWord = `Не удалось загрузить или преобразовать изображение\n`
)

func openImage(imageBytes *[]byte) (image.Image, error) {

	img, err := png.Decode(bytes.NewReader(*imageBytes))
	if err != nil {
		log.Println("Decoding error:", err.Error())
		return nil, err
	}

	return img, nil
}

func (t *Tesseract) OpenCVGreyContrast(imageBytes *[]byte) ([]byte, error) {

	img, err := openImage(imageBytes)
	if err != nil {
		log.Println("*** Error open images by bytes...")
		return []byte(""), err
	}

	//TODO: Рабочее серый
	//result := image.NewGray(img.Bounds())
	//draw.Draw(result, result.Bounds(), img, img.Bounds().Min, draw.Src)

	//TODO: Рабочий серый с контрастом
	result := imaging.Grayscale(img)
	//result = imaging.AdjustContrast(result, 10)

	//TODO: Рабочий, просто увеличение размера
	//result = imaging.Resize(img, img.Bounds().Dx()*2, img.Bounds().Dy()*2, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, result)
	if err != nil {
		log.Fatal("Decoding error:", err.Error())
		return []byte(""), err
	}
	b := buf.Bytes()
	return b, nil

	//pixels := t.CreateTensors(img)
	//newImage := t.GreyScale(pixels)
	//return t.ConvertTensor(*newImage), nil
}

func (t *Tesseract) ConvertTensor(pixels [][]color.Color) []byte {

	//Преобразование обратно из тензора
	rect := image.Rect(0, 0, len(pixels), len(pixels[0]))
	nImg := image.NewRGBA(rect)

	for x := 0; x < len(pixels); x++ {
		for y := 0; y < len(pixels[0]); y++ {
			q := pixels[x]
			if q == nil {
				continue
			}
			p := pixels[x][y]
			if p == nil {
				continue
			}
			original, ok := color.RGBAModel.Convert(p).(color.RGBA)
			if ok {
				nImg.Set(x, y, original)
			}
		}
	}

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, nImg, nil)
	if err != nil {
		fmt.Println("Decoding error:", err.Error())
		return []byte("")
	}

	b := buf.Bytes()
	return b
}

//func (t *Tesseract) MakeOCR(imageBytes []byte) (string, error) {
//
//	newImage, err := t.OpenCVGreyContrast(&imageBytes)
//	if err != nil {
//		return "", err
//	}
//
//	err = t.Client.SetImageFromBytes(newImage)
//	t.Client.SetLanguage("eng")
//	if err != nil {
//		return "", err
//	}
//	t.Client.Init()
//	t.Client.SetPageSegMode(11)
//	text, err := t.Client.Text()
//	if err != nil {
//		fmt.Println("*** Error function text", err.Error())
//		return "", err
//	}
//
//	text = strings.ReplaceAll(text, "soam", "spam")
//	text = strings.ReplaceAll(text, "Soam", "Spam")
//
//	return text, err
//}

func (p OCRProcess) CheckNumberMatch(task *Task) error {

	if task.RecognitedText == "" {
		return nil
	}

	//Checking the similarity of the number on the screenshot
	e := CheckNumberMatchOnScreen(task)
	if e == nil {
		return nil
	}
	if task.Cnam != "" {
		if CheckCnam(task.RecognitedText, task.Cnam) {
			return nil
		}
	}
	return errors.New("wrong cnam and number")
}

func CheckNumberMatchOnScreen(task *Task) error {

	reg := regexp.MustCompile(`([^0-9])`)
	newText := reg.ReplaceAllString(task.RecognitedText, "")

	result := gstr.Levenshtein(task.FromNum, newText, 1, 1, 1)

	if result <= 7 {
		return nil
	}

	number := task.FromNum[1:]
	if strings.Contains(newText, number) {
		return nil
	}
	return errors.New("wrong number")
}

func GoBash(command string, args ...string) (bt []byte, err error) {
	var scanOut *bufio.Scanner
	var stdout io.ReadCloser

	ctx, cancel := context.WithTimeout(context.Background(), 22*time.Second)
	defer func() {
		scanOut = nil
		if stdout != nil {
			stdout.Close()
		}
		cancel()
	}()
	cmd := exec.CommandContext(ctx, command, args...)

	stdout, err = cmd.StderrPipe()
	if err != nil {
		return
	}

	scanOut = bufio.NewScanner(stdout)
	err = cmd.Start()
	if err != nil {
		return
	}

	for scanOut.Scan() {
		m := scanOut.Bytes()
		bt = append(bt, m...)
	}

	err = cmd.Wait()
	if err != nil {
		return
	}

	return
}

func (p OCRProcess) NewOCR(imageName string) (string, error) {

	imageString := ""
	var err error

	for i := 0; i < 3; i++ {

		cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("/usr/bin/swift %s/ocr.swift %s/images/%s false", BasePath, BasePath, imageName))
		imageBytes, err := cmd.CombinedOutput()
		if err != nil || strings.Contains(imageString, TextStopWord) {
			outputMessage := string(imageBytes)
			p.logger.Info("swift command", zap.Any(processId, p.ProcessId), zap.Any("output", outputMessage))
			continue
		}

		imageString = string(imageBytes)
		break
	}

	if err != nil {
		return "", err
	}

	newStr := ""
	strs := strings.Split(imageString, "Recognized text:")
	if len(strs) > 1 {
		newStr = strs[0]
	}

	if newStr == "" {
		//p.logger.Info("split error", zap.Any("process id", processId.String()), zap.Any("text", string(imageBytes)))
		newStr = imageString
	}
	return newStr, nil
}

func (p *OCRProcess) StageNumberMatch() error {

	err := p.CheckNumberMatch(p.task)
	if err != nil {
		p.task.ClearImageBytes()
		p.logger.Info("stage number match", zap.Any(processId, p.ProcessId), zap.Any("task", p.task))
		p.task.NumberMatch = false
		p.logger.Info(
			"error number match",
			zap.Any(processId, p.ProcessId),
			zap.Any("number", p.task.FromNum),
			zap.Any("screenShot", p.appConfig.S3LeftLink+"/"+p.task.MessageId+"/"+p.task.Fname),
			zap.Any("cnam", p.task.Cnam),
			zap.Any("recognizedText", p.task.RecognitedText),
		)
		metric := fmt.Sprintf(`ocr_incoming_number_match{cloud_number="%s",message="false"}`, p.appConfig.Hostname)
		p.metric.SendMetric(metric, float64(1))
		return err
	}
	p.task.NumberMatch = true
	return nil
}

func (p *OCRProcess) StageOCR() error {

	var err error

	p.task.RecognitedText, err = p.NewOCR(p.task.Fname)
	if err != nil {
		p.logger.Info("stage ocr", zap.Any(processId, p.ProcessId), zap.Any("fileName", p.task.Fname))
		return err
	}
	return nil
}
