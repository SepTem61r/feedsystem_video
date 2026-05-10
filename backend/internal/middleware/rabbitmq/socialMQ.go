package rabbitmq

type SocialMQ struct {
	*RabbitMQ
}

func NewSocialMQ(base *RabbitMQ) *SocialMQ {
	return &SocialMQ{base}
}
