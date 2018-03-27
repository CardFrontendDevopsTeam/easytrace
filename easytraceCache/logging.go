package easytraceCache
import (
	"github.com/go-kit/kit/log"
	"time"
)

type loggingService struct {
	logger log.Logger
	Service
}

func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}



func (s loggingService) reloadCache() (name string) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "reloadCache",
			"name", name,
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.Service.reloadCache()
}