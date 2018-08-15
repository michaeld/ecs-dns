package lib

//Config holds the configuration for services and backends
type Config struct {
	Region, Cluster, Zone, Domain string
	Interval                      int64
}
