package http

import "github.com/gin-contrib/pprof"

// createController регистрирует публичные (mobile/Web) эндпоинты.
// Префикс /app сохранён для обратной совместимости.
func (s *Server) createController() {
	common := s.serv.Group("").Use(s.auth)

	pprof.RouteRegister(s.serv.Group(""), "debug/pprof")

	// EduGroups routes
	common.POST("/", s.withLogger(s.gzipMiddleware(s.CreateShortURL)))
	common.GET("/:id", s.withLogger(s.gzipMiddleware(s.GetByID)))
	common.GET("/ping", s.withLogger(s.gzipMiddleware(s.Ping)))

	apiGroup := s.serv.Group("").Group("/api").Use(s.auth)
	apiGroup.POST("/shorten", s.withLogger(s.gzipMiddleware(s.CreateShortURLByBody)))
	apiGroup.POST("/shorten/batch", s.withLogger(s.gzipMiddleware(s.BatchURL)))
	apiGroup.GET("/user/urls", s.withLogger(s.gzipMiddleware(s.GetUsersUrls)))
	apiGroup.DELETE("/user/urls", s.withLogger(s.gzipMiddleware(s.DeleteURLs)))
}
