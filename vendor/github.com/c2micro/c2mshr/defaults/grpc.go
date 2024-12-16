package defaults

const (
	// максимальный размер одного прото сообщения
	MaxProtobufMessageSize = 2 * 1024 * 1024 * 1024
	// название заголовка для авторизации листенера
	GrpcAuthListenerHeader = "X-Access-Token"
	// название заголовка для авторизации оператора
	GrpcAuthOperatorHeader = "Y-Access-Token"
	// название заголовка для авторизации на management сервере
	GrpcAuthManagementHeader = "Z-Access-Token"
)
