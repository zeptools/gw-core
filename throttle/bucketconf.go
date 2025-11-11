package throttle

import "time"

type BucketConf struct {
	Burst      int           // maximum number of tokens in the bucket
	Increment  int           // how many tokens to add each period
	IncrPeriod time.Duration // how often to add Increment
}
