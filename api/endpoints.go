package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pagefaultgames/pokerogue-server/api/account"
	"github.com/pagefaultgames/pokerogue-server/api/daily"
	"github.com/pagefaultgames/pokerogue-server/api/savedata"
	"github.com/pagefaultgames/pokerogue-server/defs"
	"github.com/valyala/fasthttp"
)

/*
	The caller of endpoint handler functions are responsible for extracting the necessary data from the request.
	Handler functions are responsible for checking the validity of this data and returning a result or error.
	Handlers should not return serialized JSON, instead return the struct itself.
*/

func HandleAccountInfo(ctx *fasthttp.RequestCtx) {
	username, err := usernameFromTokenHeader(string(ctx.Request.Header.Peek("Authorization")))
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}

	uuid, err := uuidFromTokenHeader(string(ctx.Request.Header.Peek("Authorization"))) // lazy
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}

	response, err := account.Info(username, uuid)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx.Response.BodyWriter()).Encode(response)
	if err != nil {
		httpError(ctx, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
}

func HandleAccountRegister(ctx *fasthttp.RequestCtx) {
	err := account.Register(string(ctx.PostArgs().Peek("username")), string(ctx.PostArgs().Peek("password")))
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(http.StatusOK)
}

func HandleAccountLogin(ctx *fasthttp.RequestCtx) {
	response, err := account.Login(string(ctx.PostArgs().Peek("username")), string(ctx.PostArgs().Peek("password")))
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx.Response.BodyWriter()).Encode(response)
	if err != nil {
		httpError(ctx, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
}

func HandleAccountLogout(ctx *fasthttp.RequestCtx) {
	token, err := base64.StdEncoding.DecodeString(string(ctx.Request.Header.Peek("Authorization")))
	if err != nil {
		httpError(ctx, fmt.Errorf("failed to decode token: %s", err), http.StatusBadRequest)
		return
	}

	err = account.Logout(token)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(http.StatusOK)
}

func HandleGamePlayerCount(ctx *fasthttp.RequestCtx) {
	ctx.SetBody([]byte(strconv.Itoa(playerCount)))
}

func HandleGameTitleStats(ctx *fasthttp.RequestCtx) {
	err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(defs.TitleStats{
		PlayerCount: playerCount,
		BattleCount: battleCount,
	})
	if err != nil {
		httpError(ctx, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
}

func HandleGameClassicSessionCount(ctx *fasthttp.RequestCtx) {
	ctx.SetBody([]byte(strconv.Itoa(classicSessionCount)))
}

func HandleSaveData(ctx *fasthttp.RequestCtx) {
	uuid, err := uuidFromTokenHeader(string(ctx.Request.Header.Peek("Authorization")))
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}

	datatype := -1
	
	if ctx.QueryArgs().Has("datatype") {
		datatype, err = strconv.Atoi(string(ctx.QueryArgs().Peek("datatype")))
		if err != nil {
			httpError(ctx, err, http.StatusBadRequest)
			return
		}
	}

	var slot int
	if ctx.QueryArgs().Has("slot") {
		slot, err = strconv.Atoi(string(ctx.QueryArgs().Peek("slot")))
		if err != nil {
			httpError(ctx, err, http.StatusBadRequest)
			return
		}
	}

	var save any
	// /savedata/get and /savedata/delete specify datatype, but don't expect data in body
	if string(ctx.Path()) != "/api/savedata/get" && string(ctx.Path()) != "/api/savedata/delete" {
		if datatype == 0 {
			var system defs.SystemSaveData
			err = json.Unmarshal(ctx.Request.Body(), &system)
			if err != nil {
				httpError(ctx, fmt.Errorf("failed to unmarshal request body: %s", err), http.StatusBadRequest)
				return
			}

			save = system
		// /savedata/clear doesn't specify datatype, it is assumed to be 1 (session)
		} else if datatype == 1 || string(ctx.Path()) == "/api/savedata/clear" {
			var session defs.SessionSaveData
			err = json.Unmarshal(ctx.Request.Body(), &session)
			if err != nil {
				httpError(ctx, fmt.Errorf("failed to unmarshal request body: %s", err), http.StatusBadRequest)
				return
			}

			save = session
		}
	}

	switch string(ctx.Path()) {
	case "/api/savedata/get":
		save, err = savedata.Get(uuid, datatype, slot)
	case "/api/savedata/update":
		err = savedata.Update(uuid, slot, save)
	case "/api/savedata/delete":
		err = savedata.Delete(uuid, datatype, slot)
	case "/api/savedata/clear":
		s, ok := save.(defs.SessionSaveData)
		if !ok {
			httpError(ctx, fmt.Errorf("save data is not type SessionSaveData"), http.StatusBadRequest)
			return
		}

		// doesn't return a save, but it works
		save, err = savedata.Clear(uuid, slot, daily.Seed(), s)
	}
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}

	if save == nil || string(ctx.Path()) == "/api/savedata/update" {
		ctx.SetStatusCode(http.StatusOK)
		return
	}

	err = json.NewEncoder(ctx.Response.BodyWriter()).Encode(save)
	if err != nil {
		httpError(ctx, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
}

func HandleDailySeed(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetBody([]byte(daily.Seed()))
}

func HandleDailyRankings(ctx *fasthttp.RequestCtx) {
	uuid, err := uuidFromTokenHeader(string(ctx.Request.Header.Peek("Authorization")))
	if err != nil {
		httpError(ctx, err, http.StatusBadRequest)
		return
	}

	var category int
	
	if ctx.QueryArgs().Has("category") {
		category, err = strconv.Atoi(string(ctx.QueryArgs().Peek("category")))
		if err != nil {
			httpError(ctx, fmt.Errorf("failed to convert category: %s", err), http.StatusBadRequest)
			return
		}
	}

	page := 1
	if ctx.QueryArgs().Has("page") {
		page, err = strconv.Atoi(string(ctx.QueryArgs().Peek("page")))
		if err != nil {
			httpError(ctx, fmt.Errorf("failed to convert page: %s", err), http.StatusBadRequest)
			return
		}
	}

	rankings, err := daily.Rankings(uuid, category, page)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx.Response.BodyWriter()).Encode(rankings)
	if err != nil {
		httpError(ctx, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
}

func HandleDailyRankingPageCount(ctx *fasthttp.RequestCtx) {
	var category int
	if ctx.QueryArgs().Has("category") {
		var err error
		category, err = strconv.Atoi(string(ctx.QueryArgs().Peek("category")))
		if err != nil {
			httpError(ctx, fmt.Errorf("failed to convert category: %s", err), http.StatusBadRequest)
			return
		}
	}

	count, err := daily.RankingPageCount(category)
	if err != nil {
		httpError(ctx, err, http.StatusInternalServerError)
		return
	}

	ctx.SetBody([]byte(strconv.Itoa(count)))
}

func httpError(ctx *fasthttp.RequestCtx, err error, code int) {
	log.Printf("%s: %s\n", ctx.Path(), err)
	ctx.Error(err.Error(), code)
}
