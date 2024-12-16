package defaults

const (
	// Максимальное количество объектов в одном чанке. Используется для разбиения выборки из БД на подмассивы (например сообщения в чате/биконы/etc)
	MaxObjectChunks int = 10000
)

const (
	// Дефолтный цвет для табличных объектов (биконы/листенеры/операторы/etc)
	DefaultColor uint32 = 0
)

const (
	// Максимальная длина имени листенера
	ListenerMaxLenName = 256

	// Максимальная длина текста заметки листенера
	ListenerMaxLenNote = 256
)

const (
	// Максимальная длина строки с матадатой ОС
	BeaconMaxLenOsMeta = 1024

	// Максимальная длина строка с hostname
	BeaconMaxLenHostname = 256

	// Максимальная длина строки с username
	BeaconMaxLenUsername = 256

	// Максимальная длина строки с domain
	BeaconMaxLenDomain = 256

	// Максимальная длина строки с именем процесса
	BeaconMaxLenProcessName = 1024

	// Максимальная длина строки с заметкой
	BeaconMaxLenNote = 256
)

const (
	// Минимальная длина сообщения в чате
	ChatMinLenMessage = 1

	// Максимальная длина сообщения в чате
	ChatMaxLenMessage = 4096

	// Имя сервера в чате (будет использоваться в качестве автора для системных нотификаций в чате)
	ChatSrvFrom = ""
)

const (
	// Максимальная длина строки с username
	CredentialMaxLenUsername = 256

	// Максимальная длина строки с password
	CredentialMaxLenPassword = 256

	// Максимальная длина строки с рилмом
	CredentialMaxLenRealm = 256

	// Максимальная длина строки с хостом
	CredentialMaxLenHost = 256

	// Максимальная длина строки с заметкой
	CredentialMaxLenNote = 256
)

const (
	// Минимальная длина строки с командой
	GroupMinCmdLen = 1
	// Максимальная длина строки с командой
	GroupMaxCmdLen = 4096
)

const (
	// Максимальная длина сообщения в таск группе
	MessageMaxLen = 4096
)
