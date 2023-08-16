package pubsub

type PubsubConfig struct {
	ProjectId string
}

func (psc *PubsubConfig) GetProjectId() string {
	// TODO: read it from a config file or env variable
	return "test"
}
