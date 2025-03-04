package config

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TunnelTestSuite struct {
	suite.Suite
}

func (s *TunnelTestSuite) TestDefaultOff() {
	httpProxy := "http://127.0.0.1:1080"
	tunnel := parseTunnel(httpProxy, httpProxy, "*:false,x.y:true")

	proxy := tunnel.getProxy("*")
	s.NotNil(proxy)
	s.Equal(proxy.String(), httpProxy)

	s.Equal(os.Getenv("NO_PROXY"), "*")

	s.Nil(tunnel.getProxy("abc.xyz"))
	proxy, err := tunnel.forcedProxy(&http.Request{
		URL: &url.URL{Host: "abc.xyz", Scheme: "https"},
	})
	s.Nil(err)
	s.NotNil(proxy)
	s.Equal(proxy.String(), httpProxy)

	s.Nil(tunnel.getProxy("x.y.z"))

	s.Nil(tunnel.getProxy("xy.z"))

	proxy = tunnel.getProxy("x.y")
	s.NotNil(proxy)
	s.Equal(proxy.String(), httpProxy)

	proxy = tunnel.getProxy("a.x.y")
	s.NotNil(proxy)
	s.Equal(proxy.String(), httpProxy)
}

func (s *TunnelTestSuite) TestDefaultOn() {
	httpProxy := "http://127.0.0.1:1080"
	tunnel := parseTunnel(httpProxy, httpProxy, "*:true,x.y:false")

	proxy := tunnel.getProxy("*")
	s.NotNil(proxy)
	s.Equal(proxy.String(), httpProxy)

	s.Equal(os.Getenv("NO_PROXY"), "")

	s.Nil(tunnel.getProxy("abc.xyz"))
	proxy, err := tunnel.autoProxy(&http.Request{
		URL: &url.URL{Host: "abc.xyz", Scheme: "https"},
	})
	s.Nil(err)
	s.NotNil(proxy)
	s.Equal(proxy.String(), httpProxy)

	proxy, err = tunnel.autoProxy(&http.Request{
		URL: &url.URL{Host: "x.y.z", Scheme: "https"},
	})
	s.Nil(err)
	s.NotNil(proxy)
	s.Equal(proxy.String(), httpProxy)

	s.Equal(tunnel.getProxy("x.y"), &url.URL{})

	s.Equal(tunnel.getProxy("a.x.y"), &url.URL{})
}

func TestTunnel(t *testing.T) {
	suite.Run(t, new(TunnelTestSuite))
}
