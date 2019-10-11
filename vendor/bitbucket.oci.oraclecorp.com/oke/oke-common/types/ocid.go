package types

const (
	OCIDEncodedEntitySizeV1 = 60 // the default size in bytes of the ocid encoded entity

	OCIDTypeCluster     = "cluster"             // the type of the ocid for a cluster
	OCIDTypeWorkRequest = "clustersworkrequest" // the type of the ocid for a workrequest for our 'clusters' service
	OCIDTypeNodePool    = "nodepool"            // the type of the ocid for a nodepool

	OCIDWorkRequestShortIDPrefix = uint8('w') // the prefix letter for the workrequest short id
)
