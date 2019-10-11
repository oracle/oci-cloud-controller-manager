package types

// ImagesV1 is the data type to display the supported images
type ImagesV1 struct {
	ImageNames []string `json:"names" yaml:"names"`
}

// ImagesRequestV1 holds the request data for getting the supported images
type ImagesRequestV1 struct {
}
