package experiment

type Config struct {
	Relations   []Relation   `json:"relations"`
	Queries     []Query      `json:"queries"`
	Encryptions []Encryption `json:"encryptions"`
}

type Relation struct {
	Name       string      `json:"name"`
	Size       int         `json:"size"`
	Attributes []Attribute `json:"attributes"`
}

type Query struct {
	Origin  string   `json:"origin"`
	Args    []string `json:"args"`
	Results []string `json:"results"`
}

type Attribute struct {
	Name       string         `json:"name"`
	Type       Type           `json:"type"`
	Constraint string         `json:"constraint"`
	Generation GenerationType `json:"generation"`
	Values     []any          `json:"values,omitempty"`
}

type Encryption struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
	Cases []Case `json:"cases"`
}

type Rule struct {
	Attribute  string         `json:"attribute"`
	Encryption EncryptionMode `json:"encryption"`
}

type Case struct {
	Transforms []Transform `json:"transforms"`
}

type Transform struct {
	Object    string `json:"object"`
	Attribute string `json:"attribute"`
	Transform string `json:"transform"`
}

type GenerationType string

const (
	FromValues    GenerationType = "from_values"
	Unique        GenerationType = "unique"
	Probabilistic GenerationType = "probabilistic"
)

type EncryptionMode string

const (
	None    EncryptionMode = "none"
	AES_CBC EncryptionMode = "aes_cbc"
	AES_GCM EncryptionMode = "aes_gcm"
	// OPE     EncryptionMode = "ope" // not implemented
)

type TransformMode string

const (
	Encrypt TransformMode = "encrypt"
	Decrypt TransformMode = "decrypt"
)

type Type string

const (
	String    Type = "string"
	Integer   Type = "integer"
	DateTime  Type = "DateTime"
	UUID      Type = "UUID"
	ByteArray Type = "bytea"
	Varchar   Type = "varchar"
	Timestamp Type = "timestamp"
)
