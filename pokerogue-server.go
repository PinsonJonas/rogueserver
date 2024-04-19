package main

import (
	"encoding/gob"
	"flag"
	"log"
	"strings"

	"github.com/pagefaultgames/pokerogue-server/api"
	"github.com/pagefaultgames/pokerogue-server/db"
	"github.com/valyala/fasthttp"
)

var serveStaticContent fasthttp.RequestHandler

func main() {
	// flag stuff
	addr := flag.String("addr", "0.0.0.0:80", "network address for api to listen on")
	wwwpath := flag.String("wwwpath", "www", "path to static content to serve")
	tlscert := flag.String("tlscert", "", "path to tls certificate to use for https")
	tlskey := flag.String("tlskey", "", "path to tls private key to use for https")

	dbuser := flag.String("dbuser", "pokerogue", "database username")
	dbpass := flag.String("dbpass", "", "database password")
	dbproto := flag.String("dbproto", "tcp", "protocol for database connection")
	dbaddr := flag.String("dbaddr", "localhost", "database address")
	dbname := flag.String("dbname", "pokeroguedb", "database name")

	flag.Parse()

	// register gob types
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

	// get database connection
	err := db.Init(*dbuser, *dbpass, *dbproto, *dbaddr, *dbname)
	if err != nil {
		log.Fatalf("failed to initialize database: %s", err)
	}

	// start web server
	serveStaticContent = fasthttp.FSHandler(*wwwpath, 0)
	
	api.Init()
	
	if *tlscert != "" && *tlskey != "" {
		err = fasthttp.ListenAndServeTLS(*addr, *tlscert, *tlskey, serve)
	} else {
		err = fasthttp.ListenAndServe(*addr, serve)
	}
	if err != nil {
		log.Fatalf("failed to create http server or server errored: %s", err)
	}
}

func serve(ctx *fasthttp.RequestCtx) {
	if strings.HasPrefix(string(ctx.Path()), "/api") {
		switch string(ctx.Path()) {
		case "/api/account/info":
			api.HandleAccountInfo(ctx)
		case "/api/account/register":
			api.HandleAccountRegister(ctx)
		case "/api/account/login":
			api.HandleAccountLogin(ctx)
		case "/api/account/logout":
			api.HandleAccountLogout(ctx)

		case "/api/game/playercount":
			api.HandleGamePlayerCount(ctx)
		case "/api/game/titlestats":
			api.HandleGameTitleStats(ctx)
		case "/api/game/classicsessioncount":
			api.HandleGameClassicSessionCount(ctx)

		case "/api/savedata/get", "/api/savedata/update", "/api/savedata/delete", "/api/savedata/clear":
			api.HandleSaveData(ctx)

		case "/api/daily/seed":
			api.HandleDailySeed(ctx)
		case "/api/daily/rankings":
			api.HandleDailyRankings(ctx)
		case "/api/daily/rankingpagecount":
			api.HandleDailyRankingPageCount(ctx)
		}

		return
	}

	serveStaticContent(ctx)
}
