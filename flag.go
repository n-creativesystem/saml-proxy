package main

// import (
// 	"flag"
// 	"os"
// 	"strings"

// 	"github.com/n-creativesystem/saml-proxy/internal/utilsconv"
// )

// type webServer struct {
// 	port int
// 	cert string
// 	key  string
// }

// var (
// 	srvConf    webServer
// 	debug      bool
// 	samlConfig string
// 	redis      string
// )

// func init() {
// 	flag.BoolVar(&debug, "debug", true, "debug mode")
// 	flag.StringVar(&samlConfig, "samlConfig", "saml.yaml", "saml configuration file name")
// 	flag.IntVar(&srvConf.port, "httpPort", 8080, "http port")
// 	flag.StringVar(&redis, "redis", "", "redis config")
// 	flag.StringVar(&srvConf.cert, "cert", "", "ssl certification file name")
// 	flag.StringVar(&srvConf.key, "key", "", "ssl key file name")
// }

// func setEnvFlag() {
// 	flag.VisitAll(func(f *flag.Flag) {
// 		name := strings.ToUpper(utilsconv.ToSnakeCase(f.Name))
// 		if s := os.Getenv(strings.ToUpper(name)); s != "" {
// 			err := f.Value.Set(s)
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	})
// }
