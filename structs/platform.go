package structs

type Platform struct {
	Name     string                  `json:"name"`
	Metadata *map[string]interface{} `json:"metadata"`
}
