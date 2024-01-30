package pipeline

import (
	"ImageProcessor/config"
	"ImageProcessor/domain"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gogf/gf/v2/text/gstr"
)

// IOCRProcess represents the interface for managing the stages of an OCR (Optical Character Recognition) process.
type IOCRProcess interface {
	// StagePullTask fetches a task for processing from a given configuration and returns the task, task identifier, and any error encountered.
	StagePullTask() (Task, string, error)
	// StageDownloadFile downloads a file associated with the given task using the provided configuration, returning the file contents and any error.
	StageDownloadFile() ([]byte, error)
	// StageRemoveFile removes a file associated with the given task using the provided configuration and returns any error encountered.
	StageRemoveFile() error
	// StageTaranUpdate performs an update operation on the Taran (an OCR engine) using the provided configuration, task, and cache, returning any error encountered.
	StageTaranUpdate(cache string) error
	// StageSaveBytesToFile saves a byte slice as a file to the specified file path and returns any error encountered.
	StageSaveBytesToFile(bytes []byte, filePath string) error
	// CheckNumberMatch check the correspondence of the number and the blues in the screenshot
	CheckNumberMatch(task Task) error
	// NewOCR starts the recognition process
	NewOCR(imageName string) (string, error)
	// RunStages runs the series of defined stages for the OCR process using the provided configuration.
	RunStages()
}

// OCRProcess represents a structure for handling OCR (Optical Character Recognition) processes.
type OCRProcess struct {
	ctx         context.Context
	logger      *zap.Logger // The logger instance for logging OCR process events and information.
	appConfig   *config.Config
	PoolChannel chan struct{}
	metric      *MetricManager
	task        *Task
	ProcessId   uint64
}

// NewOCRProcess creates a new instance of OCRProcess with the provided logger.
func NewOCRProcess(ctx context.Context, logger *zap.Logger, appConfig *config.Config, metric *MetricManager) *OCRProcess {
	ocrProcess := &OCRProcess{
		ctx:         ctx,
		logger:      logger,
		appConfig:   appConfig,
		PoolChannel: appConfig.PoolChannel,
		metric:      metric,
	}
	return ocrProcess
}

type Task struct {
	Id             uint64
	Fname          string
	Attempts       uint64
	MessageId      string
	Cnam           string
	FromNum        string
	RecognitedText string
	ImageBytes     []byte
	NumberMatch    bool
}

func (t *Task) ClearImageBytes() {
	t.ImageBytes = []byte("")
}

const (
	LevenshteinIndex = 4
	processId        = "processID"
	imagePathStorage = "/home/badmin/IM/"
	imagePathCache   = "/Users/badmin/ImageProcessor/images/"
)

func (p OCRProcess) StagePullTask() (*Task, string, error) {

	cache := p.appConfig.DCCache
	resp, err := p.appConfig.Taran.PullTask(p.appConfig.DCCache)
	if err != nil {
		return &Task{}, cache, domain.ErrStagePullTask
	}

	if resp == nil {
		return &Task{}, cache, nil
	}

	return &Task{
		Id:        resp.([]interface{})[0].(uint64),
		Fname:     resp.([]interface{})[7].(string),
		Attempts:  resp.([]interface{})[3].(uint64),
		MessageId: resp.([]interface{})[16].(string),
		FromNum:   resp.([]interface{})[11].(string),
		Cnam:      resp.([]interface{})[21].(string),
	}, cache, nil
}

func (p OCRProcess) StageDownloadFile() ([]byte, error) {
	var b []byte
	var err error
	for i := 0; i < 3; i++ {
		b, err = p.appConfig.SSHClient.Run("cat " + imagePathStorage + p.task.Fname)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		p.logger.Info("stage down file", zap.Any(processId, p.ProcessId), zap.Any("cat", imagePathStorage+p.task.Fname))
		return nil, err
	}
	return b, nil
}

func (p OCRProcess) StageRemoveFile() error {

	if p.task.Fname == "" {
		return nil
	}

	var err error

	_, err = p.appConfig.SSHClient.Run("rm " + imagePathStorage + p.task.Fname)
	if err != nil {
		return err
	}
	return err
}

func (p OCRProcess) StageTaranUpdate(cache string) error {

	fields := p.appConfig.Taran.DCCache2Fields()

	status := "dc_ready"
	incomingNumber := p.task.NumberMatch
	screenShot := p.appConfig.S3LeftLink + "/" + p.task.MessageId + "/" + p.task.Fname
	cnam := p.task.Cnam
	recognizedText := p.task.RecognitedText

	//Write text to tarantool
	_, err := p.appConfig.Taran.Conn.Update(cache, "id", []interface{}{p.task.Id}, []interface{}{
		[]interface{}{"=", fields.Status, status},
		[]interface{}{"=", fields.IncomingNumberMatch, incomingNumber},
		[]interface{}{"=", fields.TextRecognized, true},
		[]interface{}{"=", fields.Screenshot, screenShot},
		[]interface{}{"=", fields.Cnam, cnam},
		[]interface{}{"=", fields.Text, recognizedText}})
	if err != nil {
		return fmt.Errorf("error update tarantool row, when recognition: %w", err)
	}
	return nil
}

func (p OCRProcess) StageSaveBytesToFile(bytes []byte, filePath string) error {
	err := ioutil.WriteFile(filePath, bytes, 0777)
	if err != nil {
		p.logger.Info("storage save bytes to file", zap.Any(processId, p.ProcessId), zap.Any("filePath", filePath))
		return err
	}
	return nil
}

func (p OCRProcess) RunStages() {

	defer func() {
		//Free up space in the queue
		<-p.PoolChannel
	}()

	if p.ctx.Err() != nil {
		return
	}

	var err error
	var cache string

	// for metrics
	// TODO: Return later
	//var cloudNumberStr = fmt.Sprintf("%d", p.appConfig.CloudNumber)
	var cloudNumberStr = p.appConfig.Hostname

	p.logger.Debug("start pull task", zap.Any(processId, p.ProcessId))

	p.task, cache, err = p.StagePullTask()
	if err != nil {
		if strings.Contains(err.Error(), "using closed connection") {
			p.logger.Error("stage pull task failed", zap.Any(processId, p.ProcessId), zap.Error(err))
			time.Sleep(time.Second * 30)
		}
		return
	}
	if err != nil {
		p.logger.Error("pull task failed", zap.Any(processId, p.ProcessId), zap.Error(err))
		return
	}

	if p.task.Id == 0 {
		return
	}

	metricName := fmt.Sprintf(`ocr_process{cloud_number="%s"}`, cloudNumberStr)
	p.metric.SendMetric(metricName, 1)

	p.ProcessId = p.task.Id
	p.logger.Debug("stage pull task", zap.Any(processId, p.ProcessId))
	p.logger.Debug("start download file", zap.Any(processId, p.ProcessId))

	p.task.ImageBytes, err = p.StageDownloadFile()
	if err != nil {

		if strings.Contains(err.Error(), "read: operation timed out") {
			p.logger.Error("stage pull task failed", zap.Any(processId, p.ProcessId), zap.Error(err))
			time.Sleep(time.Second * 30)
		}

		p.logger.Error("download file error", zap.Any(processId, p.ProcessId), zap.Error(err))
		e := p.appConfig.Taran.ResendTask(p.task.Id)
		if e != nil {
			metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="stage download file"}`, cloudNumberStr)
			p.metric.SendMetric(metricName, 1)
			p.logger.Error("resend task error: ", zap.Any(processId, p.ProcessId), zap.Error(e))
			return
		}
		metricName = fmt.Sprintf(`ocr_retry{cloud_number="%s",reason="stage download file"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
		return
	}

	defer func(name string) {

		filepath := imagePathCache + p.task.Fname

		p.logger.Debug("remove file", zap.Any(processId, p.ProcessId), zap.Any("filepath", filepath))

		err = os.Remove(name)
		if err != nil {
			p.logger.Info("error remove file", zap.Any(processId, p.ProcessId), zap.Any("filepath", filepath))
		}
	}(imagePathCache + p.task.Fname)

	//TODO: Для теста. Потом можно просто скачивать файл, а лучше сохранять Device3 на каждый мак

	err = p.StageSaveBytesToFile(p.task.ImageBytes, imagePathCache+p.task.Fname)
	if err != nil {
		p.logger.Error("stage save to bytes to file", zap.Any(processId, p.ProcessId), zap.Error(err))
		metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="stage save to bytes to file"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
		return
	}

	p.logger.Debug("start save bytes to file", zap.Any(processId, p.ProcessId))
	p.logger.Debug("start OCR", zap.Any(processId, p.ProcessId))

	if err = p.StageOCR(); err != nil {
		p.logger.Error("stage ocr", zap.Error(err))
		if p.task.Attempts < 2 {
			if err = p.StageToS3CallFailed(p.task.ImageBytes); err != nil {
				p.logger.Error("stage to s3 call failed", zap.Any(processId, p.ProcessId), zap.Error(err))
				//TODO: Remove after debug
			}

			// TODO: Insert resend metric
			err = p.appConfig.Taran.UpdateTask(p.task.Id, []interface{}{[]interface{}{"=", 1, "resend"}, []interface{}{"=", 2, time.Now().Unix()}, []interface{}{"+", 3, 1}})
			if err != nil {
				p.logger.Error("taran update task", zap.Any(processId, p.ProcessId), zap.Error(err))
				metricName := fmt.Sprintf(`ocr_error{cloud_number="%s",message="taran update task"}`, cloudNumberStr)
				p.metric.SendMetric(metricName, 1)
				return
			}
		}
		return
	}

	if p.task.Attempts <= 2 && strings.Contains(p.task.RecognitedText, "failed to load or edit image") {
		err = p.appConfig.Taran.UpdateTask(p.task.Id, []interface{}{[]interface{}{"=", 1, "ocr_ready"}, []interface{}{"=", 2, time.Now().Unix()}, []interface{}{"+", 3, 1}})
		metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="%s"}`, cloudNumberStr, p.task.RecognitedText)
		p.metric.SendMetric(metricName, 1)
		return
	}

	if p.task.Attempts > 2 && strings.Contains(p.task.RecognitedText, "failed to load or edit image") {
		e := p.appConfig.Taran.UpdateTask(p.task.Id, []interface{}{[]interface{}{"=", 1, "resend"}, []interface{}{"=", 2, time.Now().Unix()}, []interface{}{"+", 3, 1}})
		if e != nil {
			p.logger.Error("taran update task", zap.Any(processId, p.ProcessId), zap.Error(e))
			return
		}
		return
	}

	if err = p.StageNumberMatch(); err != nil {
		p.logger.Error("stage number match", zap.Error(err))
		metricName := fmt.Sprintf(`ocr_error{cloud_number="%s",message="stage number match"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
		if p.task.Attempts < 2 {
			// TODO: Insert resend metric
			e := p.appConfig.Taran.UpdateTask(p.task.Id, []interface{}{[]interface{}{"=", 1, "resend"}, []interface{}{"=", 2, time.Now().Unix()}, []interface{}{"+", 3, 1}})
			if e != nil {
				p.logger.Error("taran update task", zap.Any(processId, p.ProcessId), zap.Error(e))
				return
			}
		}
	}

	p.logger.Debug("start OCR", zap.Any(processId, p.ProcessId))
	p.logger.Debug("start compress image", zap.Any(processId, p.ProcessId))

	compressedImageBytes, err := p.StageCompressBytes(p.task.ImageBytes)
	if err != nil {
		// fail gracefully by using original image
		p.logger.Error("stage compress bytes", zap.Any(processId, p.ProcessId), zap.Error(err))
		metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="stage compress bytes"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
		compressedImageBytes = p.task.ImageBytes
	}

	p.logger.Debug("compress image OK", zap.Any(processId, p.ProcessId))
	p.logger.Debug("stage screen to s3", zap.Any(processId, p.ProcessId))

	if err = p.StageToS3(compressedImageBytes); err != nil {
		metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="stage to s3"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
		e := p.appConfig.Taran.ResendTask(p.task.Id)
		if e != nil {
			metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="resend task"}`, cloudNumberStr)
			p.metric.SendMetric(metricName, 1)
			return
		}
		p.logger.Error("stage to s3", zap.Any(processId, p.ProcessId), zap.Error(err))
		metricName = fmt.Sprintf(`ocr_retry{cloud_number="%s",reason="stage s3"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
		return
	}

	p.logger.Debug("stage to s3 OK", zap.Any(processId, p.ProcessId))
	p.logger.Debug("stage taran update", zap.Any(processId, p.ProcessId))

	if err = p.StageTaranUpdate(cache); err != nil {
		p.logger.Error("stage taran update", zap.Error(err))
		metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="stage taran update"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)

		e := p.appConfig.Taran.ResendTask(p.task.Id)
		if e != nil {
			p.logger.Error("resend task", zap.Any(processId, p.ProcessId), zap.Any("taskId", p.task.Id))
			metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="resend task"}`, cloudNumberStr)
			p.metric.SendMetric(metricName, 1)
			return
		}
		metricName = fmt.Sprintf(`ocr_retry{cloud_number="%s",message="stage taran update"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
		return
	}

	p.logger.Debug("stage taran update OK", zap.Any(processId, p.ProcessId))

	if err = p.StageRemoveFile(); err != nil {
		p.logger.Error("stage remove file", zap.Any(processId, p.ProcessId), zap.Error(err))
		metricName = fmt.Sprintf(`ocr_error{cloud_number="%s",message="stage remove file"}`, cloudNumberStr)
		p.metric.SendMetric(metricName, 1)
	}

	metricName = fmt.Sprintf(`ocr_done{cloud_number="%s"}`, cloudNumberStr)
	p.metric.SendMetric(metricName, 1)
}

// CheckCnam Get display cnam from text and log cnams
func CheckCnam(text string, logCnams string) bool {

	a := regexp.MustCompile(`(?i)CNAM\d{1,10}:`)
	cnams := a.ReplaceAllString(logCnams, "")

	a = regexp.MustCompile(`\s{2,10}`)
	cnams = strings.TrimSpace(a.ReplaceAllString(cnams, " "))
	words := strings.Split(cnams, " ")

	var filterWords []string
	for _, v := range words {
		if len(v) < 2 {
			continue
		}
		filterWords = append(filterWords, v)
	}

	var pairs []string
	for i := 0; i < len(filterWords)-1; i++ {
		pair := strings.TrimSpace(filterWords[i]) + " " + strings.TrimSpace(filterWords[i+1])
		pairs = append(pairs, pair)
	}
	filterWords = append(filterWords, pairs...)

	//Preparing the text
	a = regexp.MustCompile(`\s{2,40}`)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.TrimSpace(a.ReplaceAllString(text, " "))

	re := regexp.MustCompile(`[^A-z0-9]`)
	textElements := re.Split(text, -1)
	//textElements := strings.Split(text, " ")

	var pairsText []string
	pairsText = append(pairsText, textElements...)

	for i := 0; i < len(textElements)-1; i++ {
		pair := strings.TrimSpace(textElements[i]) + " " + strings.TrimSpace(textElements[i+1])
		pairsText = append(pairsText, pair)
	}

	for _, fWord := range filterWords {
		for _, tWord := range pairsText {
			if len(tWord) <= LevenshteinIndex && len(fWord) <= LevenshteinIndex {
				continue
			}
			result := gstr.Levenshtein(fWord, tWord, 1, 1, 1)
			if result < LevenshteinIndex {
				return true
			}
		}

	}
	return false
}

func CheckMatchCnam(text string, logCnams string) bool {

	if logCnams == "" || logCnams == " " || text == "" || text == " " {
		return false
	}

	a := regexp.MustCompile(`(?i)CNAM\d{1}:`)
	cnams := a.Split(strings.ToLower(logCnams), -1)

	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	words := strings.Split(text, " ")
	var pairs []string
	for i := 0; i < len(words)-1; i++ {
		pairs = append(pairs, words[i]+" "+words[i+1])
	}
	words = append(words, pairs...)

	checkCnam := false
	for _, word := range words {
		for _, cnam := range cnams {
			result := gstr.Levenshtein(cnam, word, 1, 1, 1)
			if result < 4 {
				fmt.Println("===>", word, cnam, result)
				checkCnam = true
				break
			}
		}
	}

	return checkCnam
}
