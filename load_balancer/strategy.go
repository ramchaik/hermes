package loadbalancer

type Strategy interface {
	GetNextService() *Service
	UpdateServiceStats(s *Service)
}
