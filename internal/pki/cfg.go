package pki

// ConfigV1 конфигурация CA
type ConfigV1 struct {
	CA         CAConfigV1         `json:"ca"`
	Listener   ListenerConfigV1   `json:"listener"`
	Operator   OperatorConfigV1   `json:"operator"`
	Management ManagementConfigV1 `json:"management"`
}

// CAConfigV1 структура для хранения конфигурации CA
type CAConfigV1 struct {
	Serial  int64           `json:"serial" default:"1"`
	Subject SubjectConfigV1 `json:"subject"`
}

// ListenerConfigV1 структура для хранения конфигурации сертификата для сервера listener
type ListenerConfigV1 struct {
	Serial  int64           `json:"serial" default:"1"`
	Subject SubjectConfigV1 `json:"subject"`
}

// OperatorConfigV1 структура для хранения конфигурации сертификата для сервера operator
type OperatorConfigV1 struct {
	Serial  int64           `json:"serial" default:"1"`
	Subject SubjectConfigV1 `json:"subject"`
}

// ManagementConfigV1 структура для хранения конфигурации сертификата для сервера management
type ManagementConfigV1 struct {
	Serial  int64           `json:"serial" default:"1"`
	Subject SubjectConfigV1 `json:"subject"`
}

type SubjectConfigV1 struct {
	OrganizationalUnit string `json:"ou" default:"c2micro Root CA"`
	Organization       string `json:"o" default:"c2micro"`
	CommonName         string `json:"cn" default:"c2micro"`
	Country            string `json:"c" default:"RU"`
	Province           string `json:"p" default:""`
	Locality           string `json:"l" default:"Moscow"`
}
