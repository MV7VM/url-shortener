package http

// createController регистрирует публичные (mobile/Web) эндпоинты.
// Префикс /app сохранён для обратной совместимости.
func (s *Server) createController() {
	common := s.serv.Group("")

	// EduGroups routes
	common.POST("/", s.withLogger(s.CreateShortURL))
	common.GET("/:id", s.withLogger(s.GetByID))

	apiGroup := common.Group("/api")
	apiGroup.POST("/shorten", s.withLogger(s.CreateShortURLByBody))
}
