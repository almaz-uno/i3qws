package serve

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/cured-plumbum/i3qws/pkg/i3qws"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

type (
	ops struct {
		i3qws *i3qws.I3qws
	}
)

// EchoServe creates and starts echo server on Unix socket
func EchoServe(ctx context.Context, stopCh chan bool, i3qws *i3qws.I3qws, socket string) (*echo.Echo, error) {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "unix", socket)
	if err != nil {
		return nil, fmt.Errorf("is another instance server running? — %w", err)
	}

	e := echo.New()
	// e.Logger = logrus.StandardLogger()
	e.Pre(middleware.RemoveTrailingSlash())

	e.Listener = l
	o := &ops{i3qws: i3qws}
	e.Any("/focus/:num", o.focus)
	e.Any("/list", o.list)
	e.Any("/stop", o.stop(stopCh))

	go func() {
		if e2 := e.Start(""); e != nil {
			logrus.WithError(e2).Warn("Echo exited")
		}
	}()
	return e, nil
}

func (o *ops) focus(c echo.Context) error {
	snum := c.Param("num")
	num, err := strconv.Atoi(snum)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unable to convert '%s' to a number", snum)).SetInternal(err)
	}
	n, err := o.i3qws.Focus(num)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unable to focus %d window", num)).SetInternal(err)
	}
	if n == nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("There is no such window number %d", num))
	}
	return c.JSON(http.StatusOK, n)
}

func (o *ops) list(c echo.Context) error {
	nn := o.i3qws.DumpList()
	return c.JSON(http.StatusOK, nn)
}

func (o *ops) stop(stopCh chan bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		close(stopCh)
		return c.String(http.StatusOK, "done")
	}
}
