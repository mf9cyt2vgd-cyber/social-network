package repository

// Old version of kafka producer, now application use events from github.com/ThreeDotsLabs/watermill
//
//	type KafkaProducer struct {
//		*kafka.Producer
//		topic string
//	}
//
//	func NewKafkaProducer(brokers []string, topic string, log *slog.Logger) (*KafkaProducer, error) {
//		writer, err := kafka.NewProducer(&kafka.ConfigMap{
//			"bootstrap.servers": strings.Join(brokers, ","),
//		})
//		if err != nil {
//			log.Error("failed to create kafka producer", "error", err)
//			return nil, err
//		}
//		return &KafkaProducer{writer, topic}, nil
//	}
//
//	func (r *KafkaProducer) Publish(ctx context.Context, post *domain.Post) error {
//		marshalledPost, err := json.Marshal(post)
//		if err != nil {
//			return err
//		}
//		deliveryChan := make(chan kafka.Event)
//		defer close(deliveryChan)
//
//		err = r.Produce(&kafka.Message{
//			TopicPartition: kafka.TopicPartition{Topic: &r.topic, Partition: kafka.PartitionAny},
//			Value:          marshalledPost}, deliveryChan)
//		if err != nil {
//			return err
//		}
//		e := <-deliveryChan
//		m := e.(*kafka.Message)
//		if m.TopicPartition.Error != nil {
//			return fmt.Errorf("failed to deliver message: %w", m.TopicPartition.Error)
//		}
//		return nil
//	}
