package defaults

const (
	// максимальная длина имени оператора
	OperatorUsernameMaxLength = 256
	// минимальная длина имени оператора
	OperatorUsernameMinLength = 3
)

const (
	// максимальная длина имени листенера
	ListenerNameMaxLength = 256
	// максимальная длина заметки листенера
	ListenerNoteMaxLength = 256
)

const (
	// максимальная длина строки с метадатой ОС
	BeaconOsMetaMaxLength = 1024
	// максимальная длина имени хоста
	BeaconHostnameMaxLength = 256
	// максимальная длина имени пользователя
	BeaconUsernameMaxLength = 256
	// максимальная длина имени домена
	BeaconDomainMaxLength = 256
	// максимальная длина имени процесса
	BeaconProcessNameMaxLength = 1024
	// максимальная длина заметрки бикона
	BeaconNoteMaxLength = 256
)

const (
	// минимальная длина сообщения в чат
	ChatMessageMinLength = 1
	// максимальная длина сообщения в чат
	ChatMessageMaxLength = 4096
	// имя пользователя, от которого сервер публикует сообщения в чат
	ChatSrvFrom = ""
)

const (
	// максимальная длина имени пользователя в кредах
	CredentialUsernameMaxLength = 256
	// максимальная длина секрета в кредах
	CredentialSecretMaxLength = 4096
	// максимальная длина рилма в кредах
	CredentialRealmMaxLength = 256
	// максимальная длина имени хоста в кредах
	CredentialHostMaxLength = 256
	// максимальная длина заметки в кредах
	CredentialNoteMaxLength = 256
)

const (
	// минимальная длина строки с командой
	TaskGroupCmdMinLength = 1
	// максимальная длина строки с командой
	TaskGroupCmdMaxLength = 4096
	// максимальная длина сообщения в таск группе
	TaskGroupMessageMaxLength = 4096
)
