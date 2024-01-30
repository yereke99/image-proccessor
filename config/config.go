package config

import (
	"ImageProcessor/domain"
	"ImageProcessor/taran"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"os"
	"strconv"

	"github.com/melbahja/goph"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	configUrl = "http://192.168.0.73:5500/get-config"
)

type Config struct {
	AwsAccessKeyIdCid     string `yaml:"AWS_ACCESS_KEY_ID_CID"`
	AwsSecretAccessKeyCid string `yaml:"AWS_SECRET_ACCESS_KEY_CID"`

	AwsScreenBucket string `yaml:"AWS_SCREEN_BUCKET"`
	AwsAppUrl       string `yaml:"AWS_APP_URL"`
	Tarantool       string `yaml:"TARANTOOL"`
	TPassword       string `yaml:"TPASSWORD"`
	TUser           string `yaml:"TUSER"`
	DCCache         string `yaml:"TSPACE_CACHE"`
	DCCachepp       string `yaml:"DCCACHEPP"`
	S3LeftLink      string `yaml:"S3_LEFT_LINK"`

	KafkaAddr  string `yaml:"KAFKA_ADDR"`
	KafkaTopic string `yaml:"KAFKA_TOPIC"`

	Logger      *zap.Logger
	CloudNumber int64 `yaml:"CLOUD_NUMBER"`

	MetricReceiverAddress string `yaml:"METRIC_RECEIVER_ADDRESS"`

	S3          *minio.Client
	SSHClient   *goph.Client
	SSHUser     string `yaml:"SSHUser"`
	SSHPassword string `yaml:"SSHPassword"`
	SSHAddr     string `yaml:"SSHAddr"`
	SSHPort     int    `yaml:"SSHPort"`
	SSHPath     string `yaml:"SSHPath"`
	PoolChannel chan struct{}
	OCRMethod   domain.OCRMethod
	Taran       *taran.Tarantool
	Hostname    string
	RPCClient   *rpc.Client
	OCRClient   *http.Client
}

func NewConfig(logger *zap.Logger) (c *Config, err error) {
	conf := new(Config)

	conf.Logger = logger

	file, err := conf.GetConfig(configUrl)
	if err != nil {
		return nil, err
	}

	if len(file) == 0 {
		return nil, domain.ErrEmptyParams
	}

	if err := yaml.Unmarshal(file, conf); err != nil {
		return nil, err
	}

	cloudNumberOS := os.Getenv(domain.CloudNumberKey)
	if cloudNumberOS == "" {
		cloudNumberOS = "1"
	}

	cloudNumberInt, err := strconv.ParseInt(cloudNumberOS, 10, 64)
	if err != nil {
		conf.Logger.Error("error convert cloud number", zap.Error(err))
		return nil, domain.ErrConvertStringToInt
	}

	conf.CloudNumber = cloudNumberInt

	if conf.DCCachepp == "" {
		conf.DCCachepp = "DCCachePP"
	}

	conf.Taran = taran.NewTarantoolClient(conf.Tarantool, conf.TUser, conf.TPassword, logger)
	conf.Taran.DCCache = conf.DCCache

	useSSL := true
	minioClient, err := minio.New(conf.AwsAppUrl, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AwsAccessKeyIdCid, conf.AwsSecretAccessKeyCid, ""),
		Secure: useSSL,
	})

	if err != nil {
		conf.Logger.Error("minio", zap.Error(err))
	}
	conf.S3 = minioClient

	conf.SSHUser = "badmin"
	conf.SSHPassword = "fivestar2021!"
	conf.SSHAddr = "192.168.0.106"
	conf.SSHPort = 22
	conf.SSHPath = "/home/badmin/IM/"

	auth := goph.Password(conf.SSHPassword)
	if err != nil {
		conf.Logger.Error("auth error", zap.Error(err))
	}
	conf.Logger.Info("establishing SSH connection")
	conf.SSHClient, err = goph.NewUnknown(conf.SSHUser, conf.SSHAddr, auth)
	if err != nil {
		conf.Logger.Error("error when trying SSH connect", zap.Error(err))
	}

	conf.PoolChannel = make(chan struct{}, 10)

	conf.Hostname, err = os.Hostname()
	if err != nil {
		conf.Logger.Error("hostname error", zap.Error(err))
	}

	return conf, err
}

func (c *Config) GetConfig(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, domain.ErrStatusCode
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}
