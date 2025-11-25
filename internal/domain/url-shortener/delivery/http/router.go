package http

// createController регистрирует публичные (mobile/Web) эндпоинты.
// Префикс /app сохранён для обратной совместимости.
func (s *Server) createController() {
	common := s.serv.Group("")

	// EduGroups routes
	common.POST("/", s.withLogger(s.gzipMiddleware(s.CreateShortURL)))
	common.GET("/:id", s.withLogger(s.gzipMiddleware(s.GetByID)))

	apiGroup := common.Group("/api")
	apiGroup.POST("/shorten", s.withLogger(s.gzipMiddleware(s.CreateShortURLByBody)))
}
