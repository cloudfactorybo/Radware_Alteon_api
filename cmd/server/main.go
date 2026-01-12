package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"alteon-api/internal/config"
	"alteon-api/internal/handler"
	"alteon-api/internal/middleware"
	"alteon-api/internal/service"
	"alteon-api/pkg/httpclient"
)

func main() {
	// Inicializar logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Cargar configuración
	cfg := config.Load()

	logger.Infof("Configurados %d Alteon(s)", len(cfg.Alteons))

	// Inicializar cliente HTTP
	httpClient := httpclient.NewSecureClient(true)

	// Inicializar servicio multi-Alteon
	multiAlteonService := service.NewMultiAlteonService(cfg.Alteons, httpClient, logger)

	// Warmup de conexiones con Alteons (en background)
	logger.Info("Inicializando conexiones con Alteons...")
	go func() {
		time.Sleep(5 * time.Second) // Aumentado de 2 a 5 segundos
		logger.Info("Calentando service maps (esto puede tomar ~15 segundos)...")
		results, errors := multiAlteonService.GetAllServiceMaps()

		successCount := len(results)
		errorCount := len(errors)

		if successCount > 0 {
			logger.Infof("✓ Warmup completado: %d Alteon(s) listos", successCount)
		}
		if errorCount > 0 {
			logger.Warnf("✗ Warmup parcial: %d Alteon(s) con errores (reintentando en background...)", errorCount)
			// Segundo intento si el primero falla
			time.Sleep(10 * time.Second)
			results2, errors2 := multiAlteonService.GetAllServiceMaps()
			if len(results2) > 0 {
				logger.Infof("✓ Warmup reintento exitoso: %d Alteon(s) listos", len(results2))
			} else if len(errors2) > 0 {
				logger.Errorf("✗ Warmup fallido después de reintentos")
			}
		}
	}()

	// Inicializar handlers
	healthHandler := handler.NewHealthHandler()
	systemHandler := handler.NewSystemHandler(multiAlteonService, logger)
	licenseHandler := handler.NewLicenseHandler(multiAlteonService, logger)
	virtualServerHandler := handler.NewVirtualServerHandler(multiAlteonService, logger)
	monitoringHandler := handler.NewMonitoringHandler(multiAlteonService, logger)
	serviceMapHandler := handler.NewServiceMapHandler(multiAlteonService, logger)

	// Configurar router
	r := mux.NewRouter()

	// Aplicar middlewares globales
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.CORSMiddleware)

	// Rutas
	r.HandleFunc("/health", healthHandler.Health).Methods("GET")
	r.HandleFunc("/api/system", systemHandler.GetSystemInfo).Methods("GET")
	r.HandleFunc("/api/licenses", licenseHandler.GetLicenses).Methods("GET")
	r.HandleFunc("/api/virtualservers", virtualServerHandler.GetVirtualServers).Methods("GET")
	r.HandleFunc("/api/monitoring", monitoringHandler.GetMonitoring).Methods("GET")
	r.HandleFunc("/api/servicemap", serviceMapHandler.GetServiceMap).Methods("GET")

	// Configurar servidor
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor en goroutine
	go func() {
		logger.Infof("Servidor iniciado en %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Error al iniciar servidor: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Apagando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Error al apagar servidor: %v", err)
	}

	logger.Info("Servidor apagado correctamente")
}
