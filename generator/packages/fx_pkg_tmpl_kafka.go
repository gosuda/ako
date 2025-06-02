package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ako/util/module"
	"github.com/gosuda/ako/util/template"
)

func init() {
	pkgTemplateList["[MessageQueue/Kafka] sarama"] = createFxKafkaFile
}

const (
	kafkaDependency     = `github.com/IBM/sarama`
	kafkaClientTemplate = `package {{.package_name}}

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/fx"
)

const Name = "{{.client_name}}"

var Module = fx.Module("{{.package_name}}",
	fx.Provide(ConfigRegister()),
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		brokers := os.Getenv("KAFKA_{{.client_name}}_BROKERS")
		if brokers == "" {
			brokers = "localhost:9092"
		}

		const (
			ProducerMaxRetries       = 5
			ProducerFlushFrequency   = 750 * time.Millisecond
			ProducerFlushMaxMessages = 1000
			ProducerFlushBytes       = 1024 * 1024 * 10
			ProducerFlushMessages    = 1000
			ProducerMaxMessageBytes  = 1024 * 1024 * 10
			ProducerRetryBackoff     = 200 * time.Millisecond
			ProducerTimeout          = 3 * time.Second
			NetKeepAlive             = 15 * time.Second
		)

		conf := sarama.NewConfig()
		conf.Version = sarama.V2_8_0_0
		conf.ClientID = os.Getenv("KAFKA_{{.client_name}}_CLIENT_ID")
		conf.Producer.Return.Successes = true
		conf.Producer.Return.Errors = true
		conf.Producer.RequiredAcks = sarama.WaitForLocal
		conf.Producer.Retry.Max = ProducerMaxRetries
		conf.Producer.Partitioner = sarama.NewRandomPartitioner
		conf.Producer.Flush.Frequency = ProducerFlushFrequency
		conf.Producer.Flush.MaxMessages = ProducerFlushMaxMessages
		conf.Producer.Flush.Bytes = ProducerFlushBytes
		conf.Producer.Flush.Messages = ProducerFlushMessages
		conf.Producer.MaxMessageBytes = ProducerMaxMessageBytes
		conf.Producer.Retry.Backoff = ProducerRetryBackoff
		conf.Producer.Timeout = ProducerTimeout
		conf.Consumer.IsolationLevel = sarama.ReadCommitted
		conf.Consumer.Offsets.AutoCommit.Enable = false
		conf.Net.KeepAlive = NetKeepAlive

		return &Config{
			Brokers: strings.Split(brokers, ","),
			SaramaConfig:  conf,
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Brokers      []string
	SaramaConfig *sarama.Config
}

type {{.client_name}} struct {
	client sarama.Client

	consumerGroups     map[string]sarama.ConsumerGroup
	consumerGroupsLock sync.RWMutex

	producer     atomic.Pointer[sarama.SyncProducer]
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	c := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			cli, err := sarama.NewClient(param.Cfg.Brokers, param.Cfg.SaramaConfig)
			if err != nil {
				return fmt.Errorf("sarama.NewClient: %w", err)
			}

			c.client = cli

			return nil
		},
		OnStop: func(ctx context.Context) error {
			prod := c.producer.Load()
			if prod != nil {
				if err := (*prod).Close(); err != nil {
					return fmt.Errorf("producer.Close: %w", err)
				}
			}
			c.producer.Store(nil)

			c.consumerGroupsLock.Lock()
			for groupID, cg := range c.consumerGroups {
				if err := cg.Close(); err != nil {
					log.Printf("consumerGroup.Close: %v", err)
				}
				delete(c.consumerGroups, groupID)
			}
			c.consumerGroupsLock.Unlock()

			if err := c.client.Close(); err != nil {
				return fmt.Errorf("client.Close: %w", err)
			}

			return nil
		},
	})

	return c
}

type consumerGroupHandler struct {
	setup   func(session sarama.ConsumerGroupSession) error
	cleanup func(session sarama.ConsumerGroupSession) error
	consume func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
}

func newConsumerGroupHandler(setup func(session sarama.ConsumerGroupSession) error, cleanup func(session sarama.ConsumerGroupSession) error, consume func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error) *consumerGroupHandler {
	return &consumerGroupHandler{
		setup:   setup,
		cleanup: cleanup,
		consume: consume,
	}
}

func (c *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	if c.setup == nil {
		return nil
	}

	if err := c.setup(session); err != nil {
		return fmt.Errorf("consumerGroupHandler.setup: %w", err)
	}

	return nil
}

func (c *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	if c.cleanup == nil {
		return nil
	}

	if err := c.cleanup(session); err != nil {
		return fmt.Errorf("consumerGroupHandler.cleanup: %w", err)
	}

	return nil
}

func (c *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	if c.consume == nil {
		return nil
	}

	if err := c.consume(session, claim); err != nil {
		return fmt.Errorf("consumerGroupHandler.consume: %w", err)
	}

	return nil
}

func (c *{{.client_name}}) subscribe(topic string, groupID string, handler sarama.ConsumerGroupHandler) error {
	c.consumerGroupsLock.Lock()
	defer c.consumerGroupsLock.Unlock()

	if _, ok := c.consumerGroups[groupID]; ok {
		return fmt.Errorf("consumer group %s already exists", groupID)
	}

	cg, err := sarama.NewConsumerGroupFromClient(groupID, c.client)
	if err != nil {
		return fmt.Errorf("sarama.NewConsumerGroupFromClient: %w", err)
	}

	c.consumerGroups[groupID] = cg

	go func() {
		for {
			if err := cg.Consume(context.Background(), []string{topic}, handler); err != nil {
				log.Printf("consumerGroup.Consume: %v", err)
			}
		}
	}()

	return nil
}

func (c *{{.client_name}}) produce(message *sarama.ProducerMessage) (int32, int64, error) {
	prod := c.producer.Load()
	if prod == nil {
		p, err := sarama.NewSyncProducerFromClient(c.client)
		if err != nil {
			return 0, 0, fmt.Errorf("sarama.NewSyncProducerFromClient: %w", err)
		}
		c.producer.Store(&p)
		prod = &p
	}

	partition, offset, err := (*prod).SendMessage(message)
	if err != nil {
		return 0, 0, fmt.Errorf("producer.SendMessage: %w", err)
	}

	return partition, offset, nil
}
`
)

func createFxKafkaFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]
	packageName := filepath.Base(path)

	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), kafkaClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(kafkaDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
