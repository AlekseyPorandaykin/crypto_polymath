package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var logger = zap.NewNop()

func WithLogger(l *zap.Logger) {
	if l != nil {
		logger = l
	}
}

type ApplicationInfo struct {
	Application string `json:"application"`
	Author      string `json:"author"`
}

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PageHandler interface {
	RegistrationPageRoute(g *echo.Group)
}

type ApiHandler interface {
	RegistrationApiRoute(g *echo.Group)
}

type FilesHandler interface {
	RegistrationFilesRoute(g *echo.Group)
}
type ApiRouteGroup interface {
	RegistrationRouteApi(g *echo.Group)
}

type Server struct {
	info ApplicationInfo

	e *echo.Echo

	apiGroup  *echo.Group
	pageGroup *echo.Group
	fileGroup *echo.Group

	indexHtmlResponse string
}

func NewServer() *Server {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover(), middleware.CORS())
	e.File("/favicon.ico", "pkg/server/http/static/favicon.png")
	s := &Server{
		e:         e,
		apiGroup:  e.Group("/api"),
		pageGroup: e.Group(""),
		fileGroup: e.Group("/files"),
	}
	e.GET("/", s.mainPage)
	s.apiGroup.Use(errorApiMiddleware())
	s.pageGroup.Use(errorPageMiddleware())
	return s
}

func (s *Server) AddMiddleware(middleware echo.MiddlewareFunc) {
	s.e.Use(middleware)
}

func (s *Server) ApiGroup(prefix string) *echo.Group {
	if prefix != "" {
		return s.apiGroup.Group(prefix)
	}
	return s.apiGroup
}

func (s *Server) WithAuthor(author string) {
	s.info.Author = author
}
func (s *Server) WithIndexHtmlResponse(indexHtmlResponse string) {
	s.indexHtmlResponse = indexHtmlResponse
}

func (s *Server) WithApplicationName(name string) {
	s.info.Application = name
}

func (s *Server) WithFavicon(path string) {
	s.e.File("/favicon.ico", path)
}

func (s *Server) WithIndexPage(path string) {
	s.e.File("/", path)
}

func (s *Server) RegistrationPage(h PageHandler) {
	h.RegistrationPageRoute(s.pageGroup)
}

func (s *Server) RegistrationFilesHandler(h FilesHandler) {
	h.RegistrationFilesRoute(s.fileGroup)
}

func (s *Server) Run(host, port string) error {
	return s.e.Start(fmt.Sprintf("%s:%s", host, port))
}

func (s *Server) mainPage(c echo.Context) error {
	contentType := c.Request().Header.Get(echo.HeaderContentType)
	if contentType == echo.MIMEApplicationJSON {
		return c.JSON(http.StatusOK, s.info)
	}
	if s.indexHtmlResponse != "" {
		return c.HTML(http.StatusOK, s.indexHtmlResponse)
	}
	return c.String(http.StatusOK, "Index page")
}

func errorApiMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}
			httpErr, ok := err.(*echo.HTTPError)
			if ok && httpErr.Code == http.StatusNotFound {
				return c.JSON(http.StatusNotFound, nil)
			}
			logger.Error("error api http execute", zap.Error(err), zap.String("url", c.Request().URL.String()))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
	}
}

func errorPageMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}
			httpErr, ok := err.(*echo.HTTPError)
			if !ok {
				logger.Error("error http execute", zap.Error(err), zap.String("url", c.Request().URL.String()))
				return err
			}
			if httpErr.Code == http.StatusNotFound {
				return c.Redirect(http.StatusMovedPermanently, "/")
			}

			return err
		}
	}
}

func (s *Server) Close() {
	_ = s.e.Close()
}
