package main

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/rs/cors"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web/middleware"
	"github.com/vivowares/eywa/handlers"
	"github.com/vivowares/eywa/middlewares"
	"net/http"
)

func DeviceRouter() http.Handler {
	DeviceRouter := web.New()
	DeviceRouter.Use(middleware.RealIP)
	DeviceRouter.Use(middleware.RequestID)
	DeviceRouter.Use(middlewares.AccessLogging)
	DeviceRouter.Use(middleware.Recoverer)
	DeviceRouter.Use(middleware.AutomaticOptions)
	DeviceRouter.Get("/heartbeat", handlers.HeartBeatWs)
	DeviceRouter.Get("/channels/:channel_id/devices/:device_id/ws", handlers.WsHandler)
	DeviceRouter.Post("/channels/:channel_id/devices/:device_id/upload", handlers.HttpPushHandler)
	DeviceRouter.Post("/channels/:channel_id/devices/:device_id/push", handlers.HttpPushHandler)
	DeviceRouter.Get("/channels/:channel_id/devices/:device_id/poll", handlers.HttpLongPollingHandler)

	DeviceRouter.Compile()

	return DeviceRouter
}

func HttpRouter() http.Handler {
	httpRouter := web.New()
	httpRouter.Use(middleware.RealIP)
	httpRouter.Use(middleware.RequestID)
	httpRouter.Use(middlewares.AccessLogging)
	httpRouter.Use(middleware.Recoverer)
	httpRouter.Use(middleware.AutomaticOptions)

	httpRouter.Get("/heartbeat", handlers.HeartBeatHttp)
	httpRouter.Get("/greeting", handlers.Greeting)

	httpRouter.Handle("/admin/*", AdminRouter())
	httpRouter.Handle("/api/*", ApiRouter())

	fs := http.FileServer(http.Dir("static"))
	httpRouter.Handle("/*", fs)

	httpRouter.Compile()

	return httpRouter

}

func AdminRouter() http.Handler {
	admin := web.New()
	admin.Use(middleware.SubRouter)
	admin.Use(middlewares.AdminAuthenticator)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowCredentials: true,
	})
	admin.Use(c.Handler)

	admin.Get("/login", handlers.Login)

	admin.Get("/configs", handlers.GetConfig)
	admin.Put("/configs", handlers.UpdateConfig)

	admin.Get("/channels", handlers.ListChannels)
	admin.Post("/channels", handlers.CreateChannel)
	admin.Get("/channels/:id", handlers.GetChannel)
	admin.Delete("/channels/:id", handlers.DeleteChannel)
	admin.Put("/channels/:id", handlers.UpdateChannel)

	admin.Get("/dashboards", handlers.ListDashboards)
	admin.Post("/dashboards", handlers.CreateDashboard)
	admin.Get("/dashboards/:id", handlers.GetDashboard)
	admin.Delete("/dashboards/:id", handlers.DeleteDashboard)
	admin.Put("/dashboards/:id", handlers.UpdateDashboard)

	admin.Get("/channels/:id/value", handlers.QueryValue)
	admin.Get("/channels/:id/series", handlers.QuerySeries)
	admin.Get("/channels/:id/raw", handlers.QueryRaw)
	admin.Get("/channels/:id/tag_stats", handlers.GetChannelTagStats)
	admin.Get("/channels/:id/index_stats", handlers.GetChannelIndexStats)

	admin.Get("/connections/count", handlers.ConnectionCounts)
	admin.Get("/channels/:channel_id/devices/:device_id/status", handlers.ConnectionStatus)

	admin.Post("/channels/:channel_id/devices/:device_id/send", handlers.SendToDevice)
	admin.Post("/channels/:channel_id/devices/:device_id/request", handlers.RequestToDevice)

	return admin
}

func ApiRouter() http.Handler {
	api := web.New()
	api.Use(middleware.SubRouter)
	api.Use(middlewares.ApiAuthenticator)

	api.Get("/channels/:id/value", handlers.QueryValue)
	api.Get("/channels/:id/series", handlers.QuerySeries)

	api.Get("/channels/:channel_id/devices/:device_id/status", handlers.ConnectionStatus)
	api.Post("/channels/:channel_id/devices/:device_id/send", handlers.SendToDevice)
	api.Post("/channels/:channel_id/devices/:device_id/request", handlers.RequestToDevice)

	return api
}
