package types

const (
	TMCollectorPortNumber   = "4041"
	TMCollectorTokenKey     = contextKey("tm-col-token")
	TMCollectorClusterIDKey = contextKey("tm-col-cluster-id")
	TMCollectorTMIDKey      = contextKey("tm-col-tm-id")

	// TMCollectorTokenSecretKey is the key name for the token in the secret
	TMCollectorTokenSecretKey = "token"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}
