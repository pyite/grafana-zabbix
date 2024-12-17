package main

import (
	"fmt"
	"os"
	"strconv"
	"net/http"

	"github.com/alexanderzobnin/grafana-zabbix/pkg/datasource"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
)

const ZABBIX_PLUGIN_ID = "alexanderzobnin-zabbix-datasource"

func main() {
	backend.SetupPluginEnvironment(ZABBIX_PLUGIN_ID)

	pluginLogger := log.New()
	mux := http.NewServeMux()
	ds := Init(pluginLogger, mux)
	httpResourceHandler := httpadapter.New(mux)

	pluginLogger.Debug("Starting Zabbix datasource")

	//
	//  Read the buffer sizes from grafana.ini plugin section
	//
	//  Should be validated to be between 1 and 1000, default 16
	//
	maxSendSize, maxReceiveSize := LookupGRPCSizes(pluginLogger)

	err := backend.Manage(ZABBIX_PLUGIN_ID, backend.ServeOpts{
		CallResourceHandler: httpResourceHandler,
		QueryDataHandler:    ds,
		CheckHealthHandler:  ds,
		GRPCSettings: backend.GRPCSettings {
			MaxSendMsgSize: 1024 * 1024 * maxSendSize,
			MaxReceiveMsgSize: 1024 * 1024 * maxReceiveSize,
		},
	})
	if err != nil {
		pluginLogger.Error("Error starting Zabbix datasource", "error", err.Error())
	}
}

func Init(logger log.Logger, mux *http.ServeMux) *datasource.ZabbixDatasource {
	ds := datasource.NewZabbixDatasource()

	mux.HandleFunc("/", ds.RootHandler)
	mux.HandleFunc("/zabbix-api", ds.ZabbixAPIHandler)
	mux.HandleFunc("/db-connection-post", ds.DBConnectionPostProcessingHandler)

	return ds
}

func LookupGRPCSizes(logger log.Logger) (int, int) {
	//
	//  Read the buffer sizes from grafana.ini plugin section, in megabytes
	//
	//  Use 16 if undefined or out of range
	//
	//  Attempting to do proper input validation.
	//
	var maxSendSize int
	var maxReceiveSize int

	maxSendSize = 16
	maxReceiveSize = 16

        iniMaxSendSize := os.Getenv( "GF_PLUGIN_GRPC_MAX_SEND_MESSAGE_SIZE" )
        iniMaxReceiveSize := os.Getenv( "GF_PLUGIN_GRPC_MAX_RECEIVE_MESSAGE_SIZE" )

        //  Validate that they are integers between 1 and 1000, use 16 otherwise
	if v, err := strconv.Atoi(iniMaxSendSize); err == nil {
		if ( v >= 1 && v <= 1000 ) {
			maxSendSize = v
		}
	}

	if v, err := strconv.Atoi(iniMaxReceiveSize); err == nil {
		if ( v >= 1 && v <= 1000 ) {
			maxReceiveSize = v
		}
	}

	logger.Info( fmt.Sprintf( "Setting gRPC max sizes: send %d and receive %d", maxSendSize, maxReceiveSize ) )

	return maxSendSize, maxReceiveSize
}

