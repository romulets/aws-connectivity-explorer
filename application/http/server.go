package http

import (
	"asset-relations/application/controller"
	"asset-relations/support/config"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type Server struct {
	ec2Controller *controller.Ec2Controller
	logger        *slog.Logger
	cfg           config.HTTPConfig
}

func NewServer(ec2Controller *controller.Ec2Controller, logger *slog.Logger, cfg config.HTTPConfig) *Server {
	return &Server{
		ec2Controller: ec2Controller,
		logger:        logger,
		cfg:           cfg,
	}
}

func (s *Server) ListenAndServe() {
	router := http.NewServeMux()

	router.HandleFunc("GET /ec2-instances/ssh-open-to-internet", s.getInstancesOpenSSH)
	router.HandleFunc("POST /ec2-instances/fetch-graph", s.fetchInstancesGraph)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", s.cfg.Port),
		Handler: router,
	}

	s.logger.Info("Starting server on port " + s.cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		s.logger.Error("Server closed: " + err.Error())
	}
}

func (s *Server) getInstancesOpenSSH(writer http.ResponseWriter, req *http.Request) {
	partial := strings.ToLower(req.URL.Query().Get("partial")) == "true"

	res := s.ec2Controller.GetInstancesSSHOpen(req.Context(), partial)
	writer.WriteHeader(res.Status)
	s.safeWriteJson(writer, res.Content)
}

func (s *Server) fetchInstancesGraph(writer http.ResponseWriter, req *http.Request) {
	res := s.ec2Controller.FetchInstancesGraph(req.Context())
	writer.WriteHeader(res.Status)
	s.safeWriteJson(writer, res.Content)
}

func (s *Server) safeWriteJson(writer http.ResponseWriter, json []byte) {
	if json == nil {
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	if _, err := writer.Write(json); err != nil {
		s.logger.Error("HTTP Failure to write response")
	}
}
