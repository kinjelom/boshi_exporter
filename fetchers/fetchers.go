package fetchers

type Fetchers struct {
	MonitFetcher  *MonitFetcher
	SpecFetcher   *InstanceSpecFetcher
	SystemFetcher *SystemFetcher
}

func NewFetchers(boshSpecPath, monitPath string) *Fetchers {
	return &Fetchers{
		MonitFetcher:  NewMonitFetcher(monitPath),
		SpecFetcher:   NewInstanceSpecFetcher(boshSpecPath),
		SystemFetcher: NewSystemFetcher(),
	}
}
