package deploy

type US2Client struct {
	executorFactory *ExecutorFactory
}

func NewUS2Client() *US2Client {
	return &US2Client{
		executorFactory: NewExecutorFactory(),
	}
}

func (c *US2Client) GetExecutorFactory() *ExecutorFactory {
	return c.executorFactory
}
